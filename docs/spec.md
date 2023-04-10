# Ponzu(5)

_Ponzu version 0 is considered prototype at best. Be aware of this._

Ponzu is a record-based, tar-like archive format with several very specific differences:

- No Octal Strings
- Compression is a feature, not a pipe
- Deduplication as an option
- Extensions defined by version.
- Whole-Archive & File integrity checking
- Append-able: cat a b > c is valid
- Unicode compliant from the start (Yes that means 💯 Emojis are cool)
- Built for metadata included in modern filesystems (e.g. NTFS)
- Biggie Size files by default (64 Zebibytes)

## Why replace `tar(5)`?

Reading through the tar specification is a blast:

- Two kinds of tar archives (three if you consider GNU)
- The end of the file is marked with several zero-filled records
- Extensions. `pad[12]`. `devmajor`/`devminor`.
- `../../../../../../../etc/passwd` and other tarbombs

However, what makes Tar useful is that it's pretty much "read header, set idx=0, crank the read head".  
The constant string manipulation that has to be done as well as a quadtradic extraction time  
for certain GNU tar files means that there are Many ways that the format could be improved.

For an interesting look at what tar is like in practice, see

- https://mort.coffee/home/tar/
- https://invisible-island.net/autoconf/portability-tar.html
- https://superuser.com/questions/1633073/why-are-tar-xz-files-15x-smaller-when-using-pythons-tar-library-compared-to-mac
- https://mgorny.pl/articles/portability-of-tar-features.html

## What tar did right

Tar (and to a similar extent cpio) did a handful of things right:

- Fixed size blocks (tar) and straightforward design (cpio)
- Append-only record formats have their place

## How do we replace `tar(5)`

Where tar chose fixed size records of 512 bytes (the standard tape record length and block size used on most hard drives), Ponzu uses 4KiB blocks. This has three main advantages:

- Records are able to handle Extremely Long filenames, 1K UTF-8 codepoints
- Modern HDDs are aligned on 4k sectors (the "4Kn" SATA standard, ca. 2010)
- Many modern processors use 4k pages in memory


Additionally, compression should be considered free in today's age.  
Modern compression algorithms (ZStandard and Brotli, in Ponzu's case) are reaching their theoretical maximums for certain forms of data. As such, there's no decompression tradeoff for allowing compressed segments, only a cost during compression.  
Since archival is typically a one-time operation compared to the subsequent unpacking, this tradeoff is well worth it.

By considering compression a _built-in feature_, certain things can be improved upon:

- Dictionaries can be precomputed for a corpus and compression _extended_
- Varying types and qualities of compression can be applied to make efficient use of space.
- Certain files can be expanded to 10s of times their size (e.g. PNG art) and fit in 1/2 records.

By taking into consideration the source and host operating system, more information can be archived with less loss of fidelity (e.g. Windows SIDs).

# The Ponzu archive format, summary

Ponzu archives are comprised of _records_ of 4KiB chunks.

Each Ponzu record is headed by a Preamble containing:

- The characters 'PONZU\\0'.
- A one byte record type (uint8_t)
- A two-byte (uint16_t) flag field.
- A uint64_t defining the number of data segments (4K blocks) to follow
- A uint16_t defining the number of bytes used in the final data block
- A 64-byte (512 bits) BLAKE2b-512 checksum of the relevant scope (see [Checksums](#checksums))

A C implementation of the standard might use something like this:

```c
struct RECORD_PREAMBLE {
    uint8_t  magic[6];        // "PONZU\0"
    uint8_t  record_type;     // 0 = SOA, 1 = file, etc.
    uint16_t flags;           // Flag Set
    uint64_t data_len;        // # of blocks to read
    uint16_t modulo;          // # of bytes to use in last block
    uint8_t  checksum[64];    // BLAKE2b-512 checksum.
}
```

The rest of the record is an [RFC 8949 Concise Binary Object Representation (CBOR)](https://www.rfc-editor.org/rfc/rfc8949) encoded body.  

 CBOR content must not extend to greater than the 4KiB allotment. It is up to the  
implementation what information to exclude from the record to condense the information  
to 4KiB.

Compliant implementations MUST NOT allow the creation of files below the level of the prefix.

Paths (including the archive prefix) in Ponzu archives MUST be forward-relative except for symlink targets.  
A forward-relative path is a path which refers only to a child, not any sibling, cousin, or parent path.  
Examples of valid forward-relative paths include:

- `coconuts/bunches/lovely.jpg` (a prefectly reasonable path)
- `pools/../cheeses/Wensleydale.tiff` (does not go below the "current" path)
- `heads/talking/` (regular path to a directory)

Examples of invalid forward-relative paths include

- `kittens/../../dogs/puppies/newfoundland.jpg` (Creates a sibling)
- `./../bob/` (another parent directory acccess)
- `../x` (parent directory access)

A compliant implementation MAY provide a mechanism to ignore these rules, but it MUST be off by default.  
A compliant implementation MAY provide a mechanism to resolve paths within the archive and output a new, "defused"  
archive which contains no relative paths at all.  
A compliant implementation MUST default to writing only non-relative paths.

## Header Flags

The following flags are used:

| Value   | Introduced | Name               | Name                                                                 |
| ------- | ---------- | ------------------ | -------------------------------------------------------------------- |
| `0b01`  | 1          | `CONTROL_START`    | (for a control record) This is the start of an archive.              |
| `0b10`  | 1          | `CONTROL_END`      | (For a control record) This is the end of an archive.                |
| `0b100` | 1          | `CONTROL_STREAMED` | (for a control record) This archive may not contain checksums.       |
| `0b1`   | 1          | `CONTINUES`        | (For any record) This record has continuation blocks that follow it. |



Flags outside the mask of `0x00FF` are resered for implementation specific flags.

## Record Types

All Ponzu record bodies are encoded as CBOR bodies.

The defined record types are

| Value | Introduced | Name                 | Description                                                     | Length    |
| ----- | ---------- | -------------------- | --------------------------------------------------------------- | --------- |
| 0     | 1          | Control              | Start, end, or other "special" actions for the archive          | 0         |
| 1     | 1          | File                 | A regular file.                                                 | Varies    |
| 2     | 1          | Symlink              | A symbolic link to a path                                       | 0         |
| 3     | 1          | Hardlink             | A hard link to a specific inode                                 | 0         |
| 4     | 1          | Directory            | A directory                                                     | 0         |
| 5     | 1          | Zstandard Dictionary | Dictionary for ZStandard to use during decompression.           | Varies    |
| 126   | 1          | OS Special           | An OS-Special inode                                             | 0         |
| 127   | 1          | Continuation block   | Continuation of the previous record                             | Varies    |
| \>127 | 1          | Reserved             | All values > 127 are reserved for implementation defined usage. | arbitrary |

Here, length is specified as the number of data blocks after the record header.

### Archive Control (0)

An archive control record is defined by its flags:

* `CONTROL_START`: This is a start of archive record.
* `CONTROL_END`: This is the end of the archive


The Start of Archive record is used to define the paramters of an archive.

| Name    | Key | since | type   | Description                                        |
| ------- | --- | ----- | ------ | -------------------------------------------------- |
| version | 0   | 1     | Uint8  | Version of the Ponzu spec this archive conforms to |
| host    | 1   | 1     | string | Host OS type that this archive was created on      |
| prefix  | 2   | 1     | string | Prefix used by all files in this archive           |
| comment | 3   | 1     | string | Comment, text                                      |

> Note: The prefix MUST NOT begin with a leading `/` and any compliant implementation MUST discard a leading slash  
> unless the implementation gives a mechanism to "trust" the archive.

The End of Archive record 
### File

| Name            | Key | Since | type      | Description               |
| --------------- | --- | ----- | --------- | ------------------------- |
| name            | 0   | 1     | string    | filename                  |
| compressionType | 1   | 1     | integer   | Record compression type   |
| mTime           | 2   | 1     | timestamp | Modified time of the file |
| osMetadata      | 3   | 1     | map       | OS-Specific attributes    |

### Symlinks and Hardlinks

Links are Files with no data section and the following fields:

| Name       | Key | Since | type   | Description |
| ---------- | --- | ----- | ------ | ----------- |
| linkTarget | \-1 | 1     | string | Link target |

Hardlinks MUST refer to a file within the archive and MUST NOT begin with `/`.

### Directories

A directory is a File record but with a zero length and zero modulus.

### ZStandard Dictionary

a ZStandard Dictionary has no specific fields, however the following optional fields  
may be included:

| Name    | Key | Since | type   | Description                                                 |
| ------- | --- | ----- | ------ | ----------------------------------------------------------- |
| version | 0   | 1     | string | Version of ZStandard that created this dictionary, if known |

ZStandard dictionaries _must not_ be compressed.

When a Dictionary record is received, the old dictionary (if any) should be discarded.

### OS Special

For operating systems that support "Special" files (e.g. FIFOs, device nodes, etc),  
this type is used. These files generally do not contain "data".

| Name      | index | Since | type   | Description                 |
| --------- | ----- | ----- | ------ | --------------------------- |
| type      | \-1   | 1     | string | only "mknod" valid for now. |
| mknodMode | \-    | 1     | u32    | Mode for mknod              |
| mknodDev  | \-    | 1     | u32    | Dev_t value for mknod       |

### Continuation Block

A Continuation Block is specifically intended for several situations:

- Filesystems where >4GB files are not allowed, but an uncompressed >4GB file must be described
- Streamed archives where integrity must be assured during transit
- Data where it is infeasible to calculate a checksum for the full body in a reasonable amount of time

Continuation blocks have no body.

## Compression

Compression is handled on a per-file basis.

| value | Since | name      | Info                             |
| ----- | ----- | --------- | -------------------------------- |
| 0     | 1     | None      |                                  |
| 1     | 1     | ZStandard | https://facebook.github.io/zstd/ |
| 2     | 1     | Brotli    | https://github.com/google/brotli |

## Host Operating System

The following operating systems might show up:

- `unix` - A UNIX/BSD system
- `linux` - A typical Linux system
- `posix` - A POSIX-compliant system
- `winnt` - A Windows NT system, such as Windows 11
- `darwin` - A MacOS/Darwin system
- `universe` - A generic, know-nothing system.

POSIX and Linux is are supersets of UNIX.

### About the _Universe_ value:

The `universe` value is presented as a generic: Archives with the "Universe" machine are treated more or less like  
large file supporting tar archives with checksums. No file attribute metadata should be inferred or included.

## Handling archives from foreign systems and future versions.

When an implementation encounters an archive that uses an unknown or future version of the specification,  
a compliant archive utility SHOULD provide a mechanism to extract the foreign or unknown  
information alongside the data portion.

If an implementation encounters an unknown compression format or file record, it SHOULD provide a means to extract the data segment  
of the record AS-IS, writing the content to an unambiguous filename (e.g. `filename.ponzu_data`)

## Streamed Archives

Streamed archives are generated on the fly or in situations where seeking back through the file is not reasonable (e.g.  
because it is a TCP socket, TTY, etc).

Streamed archives may be comprised of precomputed file records, in which the precomputed checksum is known.  
In these cases, an individual file record may have a checksum, but a checksum of all 0 should be accepted.  


## Character encoding

All filenames in Pitch are UTF-8 encoded.

## Byte Order

All values shall be Big-Endian ("Network Order"), as defined by RFC8949.

# Security

A common vulnerability in Tar and other formats is path traversal attacks. These attacks are often  
the result of something similar to files named `../../../../../etc/sshd/authorized-keys` and the like.

Ponzu considers these paths unsafe. A Ponzu archive must only create a sub-tree.  
This may concern those who maintain package management around tar: Traditionally, package systems built around  
tar have used relative paths or paths of / to start the archive.

All Ponzu archives are given a prefix. This prefix could be interpeted as a suggestion -- e.g. an archive with  
the prefix `libgizmo-1.33.7` may be overridden with simply `libgizmo` or even ignored should the implementation  
decide to do so. Should an implementation wish, it could override the prefix with no or little ill effect.

Not described here is verifying archive authenticity or provenance. A compliant implementation may add additional  
records for such things as digital signatures. Additional, implementation-dependant keys may be added to the Start  
of Archive record to add a digial signature for the complete archive, for instance. This is not covered in version 1  
of this specification.

## Checksums

All checksums in version 1 of Ponzu are BLAKE2b-512 as defined by [RFC 7693](https://www.rfc-editor.org/rfc/rfc7693).

A record's checksum is BLAKE2B-512( preamble + CBOR + padding + body ).  
Any padding to align the end of the record to the next 4K block is ignored.  
When calculating the checksum, the preamble is zeroed out.

If a checksum is all zero, it's considered "unknown" or "uncalculated".  
Unknown checksums are not invalid -- they are simply considered unverified.

A compliant implementation SHOULD verify the contents of all data segments, even if their type is not known.  
A compliant implementation MUST verify the contents of all data segments of which their type is known.

A compliant implementation MUST NOT alter the checksum of an already checksummed segment.

# Appendix: Structures for Metadata maps

This section describes the metadata mapping used for each operating system.

All metadata entries are optional.

## Common

| Key         | index | type      | Since | Description                                                            |
| ----------- | ----- | --------- | ----- | ---------------------------------------------------------------------- |
| createdTime | \-    | timestamp | 1     | the creation time of the file                                          |
| fileSize    | \-    | uint64    | 1     | The final size on disk of the file, after reassembly and decompression |
| mimetype    | \-    | string    | 1     | If applicable, the MIME filetype                                       |
| comment     | \-    | string    | 1     | A freeform string comment                                              |

## UNIX

This encompasses most UNIX-like operating systems.

| Name  | index | Since | type                    | Description                         |
| ----- | ----- | ----- | ----------------------- | ----------------------------------- |
| owner | 0     | 1     | string                  | Owning user                         |
| group | 1     | 1     | string                  | Owning Group                        |
| mode  | 2     | 1     | uint16                  | File permissions (chmod compatible) |
| attr  | \-    | 1     | array of string         | Attributes/flags                    |
| xattr | \-    | 1     | map of string to binary | Extended Attributes                 |

## Linux

The Linux metadata contains the numbered UNIX metadata as well as the following:

| Name            | index | Since | type   | Description            |
| --------------- | ----- | ----- | ------ | ---------------------- |
| selinux_label   | \-    | 1     | string | SELinux label          |
| selinux_context | \-    | 1     | string | SELinux Context        |
| caps            | \-    | 1     | uint64 | Linux capability flags |

## POSIX

The POSIX environment contains the numbered UNIX metadata as well as

| Name | index | since | type            | Description                                   |
| ---- | ----- | ----- | --------------- | --------------------------------------------- |
| acls | \-    | 1     | Array of string | POSIX ACLs in the format described by setfacl |

The POSIX ACLs are here for historical completeness.

## WinNT

| Name       | index | type   | Description                  |
| ---------- | ----- | ------ | ---------------------------- |
| sddlString | 0     | string | SDDL ACL for the file        |
| attributes | 1     | uint16 | Windows NTFS attribute flags |

## MacOS/Darwin

The MacOS/Darwin metadata is inherited from the UNIX/BSD metadata.

# Appendix: License

This text is licensed under a Creative Commons CC BY-SA 4.0 license.  
For more information see https://creativecommons.org/licenses/by-sa/4.0/

In short:

You may adapt and share that adaptation of this standard with others,  
so long as you provide attribution and your modifications are shared under the same license.

As the Creative Commons license is not easily applicable to code, the reference implementations are under a suitable license.
