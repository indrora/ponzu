# Pitch: The new standard linear archive

pitch(1) and pitch(5) aim to replace the bad choices made in tar(1)

# pitch(5)

Pitch is a record-based, tar-like archive format with several very specific differences:

* No Octal Strings
* Compression is a feature, not a pipe
* Deduplication as an option
* No funny extensions like tar
* Whole-Archive & File integrity checking
* Append-able: cat a b > c is valid
* Unicode compliant from the start (Yes that means 💯 Emojis are cool)
* Built for metadata included in modern filesystems (e.g. NTFS)
* Biggie Size files by default (>8GiB...)

## Why replace tar(5)?

Reading through the tar specification is a blast: 

* Two kinds of tar archives (three if you consider GNU)
* The end of the file is marked with several zero-filled records
* Extensions. `pad[12]`. `devmajor`/`devminor`. 

However, what makes Tar useful is that it's pretty much "read header, set idx=0, crank the read head".
The constant string manipulation that has to be done as well as a quadtradic extraction time 
for certain GNU tar files means that there are Many ways that the format could be imroved.

## What's better?

Where tar chose fixed size records of 512 bytes (the standard tape record length and block size used on most hard drives), pitch uses 4KiB blocks. This has three main advantages:

* Records are able to handle Extremely Long filenames, 1K UTF-8 codepoints
* Modern HDDs are aligned on 4k sectors (the "4Kn" SATA standard, ca. 2010)
* Many modern processors use 4k pages in memory

4K as a block size also allows for many very useful hacks, such as short/half blocks (1,2K) and being able to hold the whole file in a single record, metadata included.

Additionally, compression should be considered free in today's age.
Modern compression algorithms (ZStandard and Brotli, in pitch's case) are reaching their theoretical maximums for certain forms of data. As such, there's no decompression tradeoff for allowing compressed segments, only a cost during compression. 
Since archival is typically a one-time operation compared to the subsequent unpacking, this tradeoff is well worth it. 

By considering compression a *built-in feature*, certain things can be improved upon:

* Dictionaries can be precomputed for a corpus and compression _extended_ 
* Varying types and qualities of compression can be applied to make efficient use of space.
* Certain files can be expanded to 10s of times their size (e.g. PNG art) and fit in 1/2 records.

By taking into consideration the source and host operating system, more information can be archived with less loss of fidelity (e.g. Windows SIDs).

# The pitch(5) format:

Pitch archives are comprised of *records*. Records come in three sizes:

* Full Size (4KiByte) -- The typical block size. 
* Half Size (2KiByte) -- For files smaller than 1...2 KiB. These actually take up
  the full block, but their content is the second 2KiB of the record. 


Throughout this document, the following constants will be referred to:

```c
#define BLOCK_SIZE 4096;        // Size that any block takes up
```

and the following types are declared:

```c
typedef timestamp_t int64_t    // A 64-bit timestamp, UNIX epoch, 1/s
```

Each pitch record has the same basic header structure:

* The first 5 bytes are the characters 'PITCH', no null
* The next byte is the record type
* The next two bytes (uint16_t) are flags. 
* The next 64 bytes (512 bits) are the SHA3-512 checksum of the relevant scope:
   - For archive header blocks, this is `blockcount` blocks, excluding the header.
   - For file header blocks, this is the SHA3-512 of `blockcount` blocks, *including padding zeros*.




```c
struct PITCH_RECORD {
    uint8_t   magic[5];      // PITCH
    uint8_t   record_type    // What kind of record is this?
    uint16_t   flags;        // flags
    uint8_t   checksum[64];  // SHA3-512 checksum of file or archive.
    union { 
        struct ARCHIVE_HEADER {
            uint16_t     version;       // 1, for now
            timestamp_t  created;       // 2000-epoch timestamp of archive creation
            uint8_t      host[24];      // What system created this archive (UNIX, NT, etc)
            uint8_t      compression;   // what compression format is used? (0=none)
            uint64_t     blockcount     // How many 4K-blocks are there to unpack?
            uint8_t      prefix[1024];   // Global prefix for all files in the archive
            uint8_t      comment[1024];  // Text comment
            uint8_t      padding[...];   // Padding to 4K
        } archive_header
        struct {
            uint8_t      path[1024];     // Filename.
            uint64_t     blockcount;     // Number of blocks the file takes up
            uint16_t     modulo;         // Number of bytes used in the last block
            uint16_t     permissions;    // UNIX-type file permissions 
            timestamp_t  ctime;          // Timestamp that this file was created on
            timestamp_t  mtime;          // Timestamp that this file was modified on
            uint8_t      attrib[1024];   // File attributes (OS-specific)
            uint8_t      padding[...];   // Pad to 4K
        } file_header_full
        struct {
            uint8_t      path[768];      // filename
            uint16_t     size;           // size
            uint16_t     permissions     // UNIX-like file permissions
            timestamp_t  ctime;          // Creation time
            timestamp_t  mtime;          // Modification Time
            uint8_t      attrib[1024]    // OS-specific attributes block
            uint8_t      padding[...];   // Padding to 2K
            uint8_t      data[2048]      // Data itself
        } file_record_half
    }


} record;
```

record types are as thus:

```c
enum RECORD_TYPE {
    HEADER,             // This is a header (at the start of an archive)
    FILE,               // This is a normal, regular file
    HARDLINK,           // This is a hardlink to a file
    SYMLINK,            // This is a symlink to a file
    DIRECTORY,          // This is a directory that should be created.
    ZDICTIONARY,        // ZStandard Dictionary 
    OS_SPECIAL = 0xFF   // This is a file that the OS knows how to make
} record_type;
```

Notable here is the lack of *UNIX-Specific* items, such as block devices, FIFOs, etc. 
We will refrain from defining too many of them here, but a UNIX system might consider
a block device an OS_SPECIAL but an NT system may have no concept of this and simply
create an empty file, or do nothing. 
Similarly, an NT system can encode NTFS Alternate Data Streams such as `foo.txt:ext:$DATA`
that encode a secondary data stream for the same file. 

1024 bytes are provided for any OS-Specific content such as ACLs, xattrs, etc. 

The header flags are as thus:

```c
enum HEADER_FLAGS {
    NONE = 0b0,
    HALF_RECORD  = 0b1,
    // other values reserved
} header_flags;
```

Compression:

```c
enum COMPRESSOR_TYPE {
    NONE = 0,
    ZSTD = 1,
    BROTLI = 2,
} compressor_type
```



## half records

With pitch archives, it is always safe to read a 4KiB block of data.
Data and headers are aligned on 4KiB offsets. When reading header blocks, the first record defines the size of headers in that block.
If the first record in the block is a half record, the second half of the block will either be the contents of a file OR a complete record. 

