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

For an interesting look at what tar is like in practice, see

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

The rest of the record is an [RFC 8949 Concise Binary Object Representation (CBOR)](https://www.rfc-editor.org/rfc/rfc8949) encoded body.
If the `HALF_RECORD` flag is used, the CBOR content extends only into the first 2048 bytes.
The CBOR content must not extend to greater than the 4KiB allotment. It is up to the
implementation what information to exclude from the record to condense the information
to 4KiB.

## Header Flags

The following flags are used:

| Value   | Introduced | Name               | Name                                                                              |
| ------- | ---------- | ------------------ | --------------------------------------------------------------------------------- |
| `0b1`   | 1          | `HALF_RECORD`      | The second half of the 4KiB block is the data portion                             |
| `0b10`  | 1          | `STREAMED_ARCHIVE` | (for an SOA record) This archive was streamed on the fly                          |
| `0b10`  | 1          | `NO_CHECKSUM`      | (for any other record) This files checksum could not be computed during streaming |
| `0b100` | 1          | `STAMPED`          | This record has been postfacto checksummed                                        |

All flags above `0x7F` are reserved for use by implementations. 

## Record Types

All Ponzu record bodies are encoded as CBOR bodies.

The defined record types are

| Value | Introduced | Name                 | Description                                                     |
| ----- | ---------- | -------------------- | --------------------------------------------------------------- |
| 0     | 1          | SOA                  | Start of Archive: indicates new archive context parameters.     |
| 1     | 1          | File                 | A regular file.                                                 |
| 2     | 1          | Symlink              | A symbolic link to a path                                       |
| 3     | 1          | Hardlink             | A hard link to a specific inode                                 |
| 4     | 1          | Directory            | A directory                                                     |
| 5     | 1          | Zstandard Dictionary | Dictionary for ZStandard to use during decompression.           |
| 127   | 1          | OS Special           | An OS-Special inode                                             |
| >127  | 1          | Reserved             | All values > 127 are reserved for implementation defined usage. |

### Start Of Archive (0)

The Start of Archive record is used to define the paramters of an archive. 

| Name    | Key | since | type   | Description                                        |
| ------- | --- | ----- | ------ | -------------------------------------------------- |
| version | 0   | 1     | Uint8  | Version of the Ponzu spec this archive conforms to |
| host    | 1   | 1     | string | Host OS type that this archive was created on      |
| prefix  | 2   | 1     | string | Prefix used by all files in this archive           |
| comment | 3   | 1     | string | Comment, text                                      |

Optionally, the following fields might appear:

| Name   | index | since | type | Description                                         |
| ------ | ----- | ----- | ---- | --------------------------------------------------- |
| uidmap | -1    | 1     | map  | Mapping of UID numbers to user names (e.g. for NFS) |


### File

| Name            | Key | Since | type      | Description                         |
| --------------- | --- | ----- | --------- | ----------------------------------- |
| name            | 0   | 1     | string    | filename                            |
| mode            | 1   | 1     | uint16    | File permissions (chmod compatible) |
| owner           | 2   | 1     | string    | Owning user                         |
| group           | 3   | 1     | string    | Owning Group                        |
| mTime           | 4   | 1     | timestamp | Modified time of the file           |
| compressionType | 5   | 1     | integer   | Record compression type             |
| osMetadata      | 6   | 1     | map       | OS-Specific attributes              |

### Symlinks and Hardlinks

Links are Files with no data section and the following fields:

| Name       | Key | Since | type   | Description |
| ---------- | --- | ----- | ------ | ----------- |
| linkTarget | -1  | 1     | string | Link target |

### Directories

A directory is a File record but with a zero length and zero modulus.


### ZStandard Dictionary

a ZStandard Dictionary has no specific fields, however the following optional fields
may be included:

| Name    | Key | Since | type   | Description                                                 |
| ------- | --- | ----- | ------ | ----------------------------------------------------------- |
| version | 0   | 1     | string | Version of ZStandard that created this dictionary, if known |

ZStandard dictionaries *must not* be compressed. 

When a Dictionary record is received, the old dictionary (if any) should be discarded.

### OS Special

For operating systems that support "Special" files (e.g. FIFOs, device nodes, etc),
this type is used. These files generally do not contain "data". 

| Name      | index | Since | type   | Description                 |
| --------- | ----- | ----- | ------ | --------------------------- |
| type      | -1    | 1     | string | only "mknod" valid for now. |
| mknodMode | -     | 1     | u32    | Mode for mknod              |
| mknodDev  | -     | 1     | u32    | Dev_t value for mknod       |


## Compression

Compression is handled on a per-file basis.

| value | Since | name      | Info                             |
| ----- | ----- | --------- | -------------------------------- |
| 0     | 1     | None      |                                  |
| 1     | 1     | ZStandard | https://facebook.github.io/zstd/ |
| 2     | 1     | Brotli    | https://github.com/google/brotli |


## Host Operating System 

The following operating systems might show up:

* `linux` - A typical Linux system
* `unix`  - A UNIX/BSD system
* `winnt` - A Windows NT system, such as Windows 11
* `darwin` - A MacOS/Darwin system
* `universe` - A generic, know-nothing system.

### About the _Universe_ value: 

The `universe` value is presented as a generic: Archives with the "Universe" machine are treated more or less like
large file supporting tar archives with checksums. No file attribute metadata should be inferred.

## Handling archives from foreign systems and future versions.

When an implementation encounters an archive that uses an unknown or unexpressable
a compliant archive utility SHOULD provide a mechanism to extract the foreign or unknown
information alongside the data portion. 

If an implementation encounters an unknown compression format or file record, it SHOULD extract the data segment
of the record AS-IS, writing the content to an unambiguous filename (e.g. `filename.ponzu_data`)

A compliant *library* implementation MUST provide a way to inspect the archive record itself, including CBOR data,
whenever the user wants it. 

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

All values shall be Big-Endian ("Network Order"), as defined by RFC8949.

## Checksums

All checksums in version 1 of Ponzu are SHA3-512 as defined by [FIPS PUB 202](https://csrc.nist.gov/publications/detail/fips/202/final).

If a checksum is all zero, it's considered "unknown" or "uncalculated".
Unknown checksums are not invalid -- they are simply considered unreliable. 

An archive may have its contents postfacto checked. 

The checksum for a complete archive is made by skipping the first 4K (the archive header) and computing the checksum
of the remaining 
The checksum for a file's contents includes any padding used to align to 4K blocks; that is, all checks
are done against the full, padded data of the file.

# Appendix: Structures for Metadata maps

This section describes the metadata mapping used for each operating system.

All metadata entries are optional. 
## Common

| Key   | index | type | Since | Description         |
| ----- | ----- | ---- | ----- | ------------------- |
| xattr | -     | map  | 1     | Extended Attributes |



## Linux

| Name            | index | Since | type                   | Description     |
| --------------- | ----- | ----- | ---------------------- | --------------- |
| selinux_label   | -     | 1     | string                 | SELinux label   |
| selinux_context | -     | 1     | string                 | SELinux Context |
| caps            | -     | 1     | uint64 |Linux capability flags |


## UNIX/BSD

Nothing special for UNIX/BSD

## WinNT

| Name       | index | type   | Description           |
| ---------- | ----- | ------ | --------------------- |
| sddlString | -     | string | SDDL ACL for the file |


## MacOS

Nothing special for MacOS...

# Appendix: License

This text is licensed under a Creative Commons CC BY-SA 4.0 license.
For more information see https://creativecommons.org/licenses/by-sa/4.0/

In short:

You may adapt and share that adaptation of this standard with others,
so long as you provide attribution and your modifications are shared under the same license.

As the Creative Commons license is not easily applicable to code, the reference implementations are under a suitable license.
