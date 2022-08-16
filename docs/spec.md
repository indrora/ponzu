# Ponzu(5)

_Ponzu version 0 is considered prototype at best. Be aware of this._

Ponzu is a record-based, tar-like archive format with several very specific differences:

* No Octal Strings
* Compression is a feature, not a pipe
* Deduplication as an option
* Extensions defined by version.
* Whole-Archive & File integrity checking
* Append-able: cat a b > c is valid
* Unicode compliant from the start (Yes that means 💯 Emojis are cool)
* Built for metadata included in modern filesystems (e.g. NTFS)
* Biggie Size files by default (64 Zebibytes)

## Why replace `tar(5)`?

Reading through the tar specification is a blast: 

* Two kinds of tar archives (three if you consider GNU)
* The end of the file is marked with several zero-filled records
* Extensions. `pad[12]`. `devmajor`/`devminor`. 

However, what makes Tar useful is that it's pretty much "read header, set idx=0, crank the read head".
The constant string manipulation that has to be done as well as a quadtradic extraction time 
for certain GNU tar files means that there are Many ways that the format could be imroved.

There's other reasons to drop tar. For an interesting look at what tar is like in practice, see

* https://mort.coffee/home/tar/
* https://invisible-island.net/autoconf/portability-tar.html
* https://superuser.com/questions/1633073/why-are-tar-xz-files-15x-smaller-when-using-pythons-tar-library-compared-to-mac
* https://mgorny.pl/articles/portability-of-tar-features.html

## What tar did right

Tar (and to a similar extent cpio) did a handful of things right:

* Fixed size blocks (tar) and straightforward design (cpio)
* Append-only record formats have their place 

## How do we replace `tar(5)`

Where tar chose fixed size records of 512 bytes (the standard tape record length and block size used on most hard drives), Ponzu uses 4KiB blocks. This has three main advantages:

* Records are able to handle Extremely Long filenames, 1K UTF-8 codepoints
* Modern HDDs are aligned on 4k sectors (the "4Kn" SATA standard, ca. 2010)
* Many modern processors use 4k pages in memory

4K as a block size also allows for many very useful hacks, such as short/half blocks (1,2K) and being able to hold the whole file in a single record, metadata included.

Additionally, compression should be considered free in today's age.
Modern compression algorithms (ZStandard and Brotli, in Ponzu's case) are reaching their theoretical maximums for certain forms of data. As such, there's no decompression tradeoff for allowing compressed segments, only a cost during compression. 
Since archival is typically a one-time operation compared to the subsequent unpacking, this tradeoff is well worth it. 

By considering compression a *built-in feature*, certain things can be improved upon:

* Dictionaries can be precomputed for a corpus and compression _extended_
* Varying types and qualities of compression can be applied to make efficient use of space.
* Certain files can be expanded to 10s of times their size (e.g. PNG art) and fit in 1/2 records.

By taking into consideration the source and host operating system, more information can be archived with less loss of fidelity (e.g. Windows SIDs).

# The Ponzu(5) format:

Ponzu archives are comprised of *records*. Records come in two sizes:

* Full Size (4KiByte) -- The typical block size. 
* Half Size (2KiByte) -- For files smaller than 1...2 KiB. These actually take up
  the full block, but their content is the second 2KiB of the record. 

Each Ponzu record is headed by a Preamble containing:

* The characters 'PONZU\0'. 
* A one byte record type (uint8_t)
* A two-byte (uint16_t) flag field. 
* A uint64_t defining the length of the data segment to follow
* A uint16_t defining the number of bytes used in the final data segment
* A 64-byte (512 bits) SHA3-512 checksum of the relevant scope:
   - For a Start of Archive record, the bytes following the SOA record until the next SOA record is read
   - For all other records, the content of the data segment of the record. 

A C implementation of the standard might use something like this:

```c
struct RECORD_PREAMBLE {
    uint8_t  magic[6];        // "PONZU\0"
    uint8_t  record_type;     // 0 = SOA, 1 = file, etc.
    uint64_t data_len;        // # of blocks to read
    uint16_t modulo;          // # of bytes to use in last block
    uint8_t  checksum[64];    // SHA3-512 checksum.
}
```

The rest of the record is a CBOR encoded body.
If the `HALF_RECORD` flag is used, the CBOR content extends only into the first 2048 bytes.
The CBOR content must not extend to greater than the 4KiB allotment. It is up to the
implementation what information to exclude from the record to condense the information
to 4KiB.

## Header Flags

The following flags are used:

| flag | Name | Description|
| 1 | `HALF_RECORD` | The second half of the 4KiB block is the data portion |
| 2 | `STREAMED_ARCHIVE` (for an SOA record) | This archive was streamed on the fly |
| 2 | `NO_CHECKSUM` (for any other record) | This files checksum could not be computed during streaming |

All flags above `0x7F` are reserved for use by implementations. 


```c
enum HEADER_FLAGS {
    NONE = 0b0,
    HALF_RECORD  = 0b1,
    STREAMED_ARCHIVE = 0b10,
    NO_CHECKSUM      = 0b10,
    // other values reserved
} header_flags;
```

## Record Types

All Ponzu record bodies are encoded as CBOR bodies.

Several 


```c
enum RECORD_TYPE {
    SOA = 0,         // This is a header (at the start of an archive)
    FILE = 1,               // This is a normal, regular file
    SYMLINK = 2,            // This is a symlink to a file
    HARDLINK = 3,           // This is a hardlink to a file
    DIRECTORY = 4,          // This is a directory that should be created.
    ZSTD_DICT = 5,        // ZStandard Dictionary 
    OS_SPECIAL = 0x7F   // This is a file that the OS knows how to make
    RESERVED = 0xF0;    // All values above 127 are "reserved" for implementations.
} rtype;
```

### Start Of Archive (0)

The Start of Archive record is used to define the paramters of an archive. 

| index | type | Description |
|-------|------|----------------------------|
| 0     | Uint8 | Version of the Ponzu spec this archive conforms to |
| 1     | string | Host OS type that this archive was created on |
| 2     | uint8 | Compression type used in records |
| 3     | string | Prefix used by all files in this archive |
| 4     | string  | Comment, text |

### File

| index | type      | Description               |
|-------|-----------|---------------------------|
| 0     | string    | filename                  |
| 1     | uint16    | File permissions          |
| 2     | timestamp | Modified time of the file |
| 3     | map       | OS-Specific attributes    |

### Symlink

Symlinks use all the same indexes as a file, adding

| index | type | Description | 
| 4 | string | Link target |

### Hardlink 

Hardlinks use the same structure as a Symlink.

### Directories

A Directory has all the fields of a File

### ZStandard Dictionary

a ZStandard Dictionary has no specific fields, however the following optional fields
may be included:

| index | type | Description |
| 0 | string | Version of ZStandard that created this dictionary, if known |

### OS Special

For operating systems that support "Special" files (e.g. FIFOs, device nodes, etc),
this type is used. These files generally do not contain "data". 

| index | type | Description |
| 0 | string | "mknod" or other. |
| 10 | u32 | Mode for mknod |
| 11 | u32 | Dev_t value for mknod |



## Compression

Compression is handled on a per-file basis. 

```c
enum COMPRESSOR_TYPE {
    NONE = 0,
    ZSTD = 1,
    BROTLI = 2,
} compressor_type
```

## Host Operating System 

host operating systems:
```c
#define HOST_OS_LINUX = "linux"
#define HOST_OS_UNIX  = "unix"
#define HOST_OS_SELINUX = "selinux"
#define HOST_OS_NT ="winnt"
#define HOST_OS_GENERIC = "universe"
#define HOST_OS_DARWIN  = "darwin"
```

### About the _Universe_ value: 

The `universe` value is presented as a generic: Archives with the "Universe" machine are treated more or less like
large file supporting tar archives with checksums. No file attribute metadata should be inferred.

## Streamed Archives

Streamed archives are special. Astute readers will notice that NO_CHECKSUM and STREAMED_ARCHIVE have the same value.
Streamed archives will not have the time to know ahead of time what the checksum of the compressed data is, how long
the archive is going to be, etc. 

Streamed archives may be comprised of precomputed file records, in which the precomputed checksum may be already known.
In these cases, an individual file record may have a checksum.
A streamed archive may be verified post-facto and have its checksums "stamped" upon it. 

## Character encoding

All filenames in Pitch are UTF-8 encoded. 

## Byte Order

All values shall be Big-Endian ("Network Order"). 

## Checksums

All checksums in version 1 of Pitch are SHA3-512.
If a checksum is all zero, it's considered "unknown" or "uncalculated".
Unkown checksums are not invalid -- they are simply considered unreliable. 
Some types, such as symlinks, have no checksum (how does one checksum a symlink?)

An archive may have its contents postfacto checked.

The checksum for a complete archive is made by skipping the first 4K (the archive header) and computing the checksum
of the remaining 
The checksum for a file's contents includes any padding used to align to 4K blocks; that is, all checks
are done against the full, padded data of the file.


# License

This text is licensed under a Creative Commons CC BY-SA 4.0 license.
For more information see https://creativecommons.org/licenses/by-sa/4.0/

In short:

You may adapt and share that adaptation of this standard with others,
so long as you provide attribution and your modifications are shared under the same license.

As the Creative Commons license is not easily applicable to code, the reference implementations are under a suitable license.
