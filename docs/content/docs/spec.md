---
weight: 200
title: "The Ponzu Spec"
description: ""
icon: "article"
date: "2023-12-19T22:26:17-08:00"
lastmod: "2023-12-19T22:26:17-08:00"
draft: false
toc: true
---

This document outlines the specification for the Ponzu archive format. 

# Introduction

Ponzu archives are comprised of *records* of 4KiB chunks. 

Each Ponzu record is headed by a Preamble containing:

- The characters ‘PONZU\0’ - including the null termination.
- A one byte record type (uint8_t)
- A one-byte compression type
- A two-byte (uint16_t) flag field.
- A uint64_t defining the number of data segments (4K blocks) to follow
- A uint16_t defining the number of bytes used in the final data block
- A 64-byte (512-bit) BLAKE2b-512 checksum of the metadata section
- A uint16_t defining the length of the metadata section
- A 64-byte (512 bits) BLAKE2b-512 checksum of content

A C implementation of the standard might use something like this:

```c
struct RECORD_PREAMBLE {
    uint8_t  magic[6];        // "PONZU\0"
    uint8_t  record_type;     // 0 = SOA, 1 = file, etc.
    uint8_t  compression;     // Type of compression
    uint16_t flags;           // Flag Set
    uint64_t data_len;        // # of blocks to read
    uint16_t data_modulo;          // # of bytes to use in last block
    uint8_t  data_checksum[64];    // BLAKE2b-512 of the data blocks to follow.
		uint16_t metadata_length;      // length of the metadata to be read
		uint8_t  metadata_checksum[64];// BLAKE2b-512 checksum of the metadata 
}
```

This preamble is followed by an [RFC 8949 Concise Binary Object Representation (CBOR)](https://www.rfc-editor.org/rfc/rfc8949) encoded body of metadata, padded to the nearest 4KiB, followed by a series of 4KiB data chunks.

*A previous version of this spec did not allow the metadata to extend past the 4KiB boundary. now it simply must be padded to the nearest 4KiB boundary.*


![record layout](images/record_format.svg)

Each record looks like this: 

```goat 
                              |<-- modulo  -->|
.--------.-------- ~ --.----------------------.-- ~ ----.
| header | Metadata    |  Body data           | padding |
'--------'-------- ~ --'----------------------'-- ~ ----'
.<-- padded to 4KiB --> <--    block_count x 4KiB    -->.
```

# Paths

Paths (including the archive prefix) in Ponzu archives MUST be forward-relative except for symlink targets.
A forward-relative path is a path which refers only to a child, not any sibling, cousin, or parent path.
Examples of valid forward-relative paths include:

- `coconuts/bunches/lovely.jpg` (a perfectly reasonable path)
- `pools/../cheeses/Wensleydale.tiff` (does not go below the “current” path)
- `heads/talking/` (regular path to a directory)

Examples of invalid forward-relative paths include

- `kittens/../../dogs/puppies/newfoundland.jpg` (Creates a sibling)
- `./../bob/` (another parent directory access)
- `../x` (parent directory access)

{{< alert icon="" context="info" >}}
Compliant implementations MUST NOT allow the creation of files below the level of the prefix.
{{< /alert >}}

{{< alert icon="" context="info" >}}
A compliant implementation MAY provide a mechanism to ignore these rules, but it MUST be off by default.
{{< /alert >}}
{{< alert icon="" context="info" >}}
A compliant implementation MAY provide a mechanism to resolve paths within the archive and output a new, “defused” archive which contains no relative paths at all.
{{< /alert >}}
{{< alert icon="" context="info" >}}
A compliant implementation MUST default to writing only non-relative paths.
{{< /alert >}}

# The most minimal Ponzu archive

The most minimal Ponzu archive consists purely of two control records: a `CONTROL_START` record and a `CONTROL_END` record. 

# Header Flags

The following flags are used:

| Value  | Introduced | Name             | Name                                                                 | Context |
| ------ | ---------- | ---------------- | -------------------------------------------------------------------- | ------- |
| 0b0001 | 1          | CONTROL_START    | (for a control record) This is the start of an archive.              | Control |
| 0b0010 | 1          | CONTROL_END      | (For a control record) This is the end of an archive.                | Control |
| 0b0100 | 1          | CONTROL_STREAMED | (for a control record) This archive may not contain checksums.       | Control |
| 0b0001 | 1          | CONTINUES        | (For any record) This record has continuation blocks that follow it. | Any     |

Flags outside the mask of `0x00FF` are reserved for implementation specific flags.

# Record Types

All Ponzu record headers are encoded as CBOR bodies.

The defined record types are

| Type value | Introduced | Name                 | Description                                                     | Data chunk count |
| ---------- | ---------- | -------------------- | --------------------------------------------------------------- | ---------------- |
| 0          | 1          | Control              | Start, end, or other “special” actions for the archive          | 0                |
| 1          | 1          | File                 | A regular file.                                                 | Varies           |
| 2          | 1          | Symlink              | A symbolic link to a path                                       | 0                |
| 3          | 1          | Hardlink             | A hard link to a specific inode                                 | 0                |
| 4          | 1          | Directory            | A directory                                                     | 0                |
| 5          | 1          | Zstandard Dictionary | Dictionary for ZStandard to use during decompression.           | Varies           |
| 126        | 1          | OS Special           | An OS-Special inode                                             | 0                |
| 127        | 1          | Continuation block   | Continuation of the previous record                             | Varies           |
| >127       | 1          | Reserved             | All values > 127 are reserved for implementation defined usage. | arbitrary        |

Here, length is specified as the number of data blocks after the record header.

## Archive Control (0)

An archive control record is defined by its flags:

- `CONTROL_START`: This is a start of archive record.
- `CONTROL_END`: This is the end of the archive

The Start of Archive record is used to define the paramters of an archive.

| Name    | Key | since | type   | Description                                        |
| ------- | --- | ----- | ------ | -------------------------------------------------- |
| version | 0   | 1     | Uint8  | Version of the Ponzu spec this archive conforms to |
| host    | 1   | 1     | string | Host OS type that this archive was created on      |
| prefix  | 2   | 1     | string | Prefix used by all files in this archive           |
| comment | 3   | 1     | string | Comment, text                                      |

{{< alert icon="" context="info" >}}
💡 Note: The prefix MUST NOT begin with a leading / and any compliant implementation MUST discard a leading slashunless the implementation gives a mechanism to “trust” the archive.

{{< /alert >}}

The End of Archive record is simply a marker that the end of the archive has been achieved. 

## File

| Name       | Key | Since | type      | Description               |
| ---------- | --- | ----- | --------- | ------------------------- |
| name       | 0   | 1     | string    | filename                  |
| mTime      | 1   | 1     | timestamp | Modified time of the file |
| osMetadata | 2   | 1     | map       | OS-Specific attributes    |

## Symlinks and Hardlinks

Links are Files with no data section and the following fields:

| Name       | Key | Since | type   | Description |
| ---------- | --- | ----- | ------ | ----------- |
| linkTarget | -1  | 1     | string | Link target |

Hardlinks MUST refer to a file within the archive and MUST NOT begin with `/`.

## Directories

A directory is a File record but with a zero length and zero modulus.

## ZStandard Dictionary

a ZStandard Dictionary has no specific fields, however the following optional fields
may be included:

| Name    | Key | Since | type   | Description                                                 |
| ------- | --- | ----- | ------ | ----------------------------------------------------------- |
| version | 0   | 1     | string | Version of ZStandard that created this dictionary, if known |

ZStandard dictionaries *must not* be compressed.

When a Dictionary record is received, the old dictionary (if any) should be discarded.

## OS Special

For operating systems that support “Special” files (e.g. FIFOs, device nodes, etc),
this type is used. These files generally do not contain “data”.

| Name      | index | Since | type   | Description                 |
| --------- | ----- | ----- | ------ | --------------------------- |
| type      | -1    | 1     | string | only “mknod” valid for now. |
| mknodMode | -     | 1     | u32    | Mode for mknod              |
| mknodDev  | -     | 1     | u32    | Dev_t value for mknod       |

## Continuation Block

A Continuation Block is specifically intended for several situations:

- Filesystems where >4GB files are not allowed, but an uncompressed >4GB file must be described
- Streamed archives where integrity must be assured during transit
- Data where it is infeasible to calculate a checksum for the full body in a reasonable amount of time

Continuation blocks have no body.


# Details of implementation

This section outlines specific details about the format, such as types of compression, byte order, and a discussion on security.

## Compression

Two algorithms are defined for compression in Ponzu: ZStandard and Brotli. Compression is only applied to data blocks that follow the record header. 

| value | Since | name      | Info                             |
| ----- | ----- | --------- | -------------------------------- |
| 0     | 1     | None      |                                  |
| 1     | 1     | ZStandard | https://facebook.github.io/zstd/ |
| 2     | 1     | Brotli    | https://github.com/google/brotli |

Compression is applied only to the data chunks that follow a record header. 

## Host Operating System values

The following operating systems might show up:

- `unix` - A UNIX/BSD system
- `linux` - A typical Linux system
- `posix` - A POSIX-compliant system
- `winnt` - A Windows NT system, such as Windows 11
- `darwin` - A MacOS/Darwin system
- `universe` - A generic, know-nothing system.

POSIX and Linux is are supersets of UNIX.

### The `universe` host

The `universe` value is presented as a generic: Archives with the “Universe” machine are treated more or less like large file supporting tar archives with checksums. No file attribute metadata should be inferred or included.

## Handling archives from foreign systems and future versions.

When an implementation encounters an archive that uses an unknown or future version of the specification, a compliant archive utility SHOULD provide a mechanism to extract the foreign or unknown information alongside the data portion.

If an implementation encounters an unknown compression format or file record, it SHOULD provide a means to extract the data segment of the record AS-IS, writing the content to an unambiguous filename (e.g. `filename.ponzu_data`)

## Streamed Archives

Streamed archives are generated on the fly or in situations where seeking back through the file is not reasonable (e.g. because it is a TCP socket, TTY, etc).

Streamed archives may be comprised of precomputed file records, in which the precomputed checksum is known. In these cases, an individual file record may have a checksum, but a checksum of all 0 should be accepted.

## Character encoding

All filenames in Pitch are UTF-8 encoded.

## Byte Order

All values shall be Big-Endian (“Network Order”), as defined by RFC8949.

## Security

A common vulnerability in Tar and other formats is path traversal attacks. These attacks are often
the result of something similar to files named `../../../../../etc/sshd/authorized-keys` and the like.

Ponzu considers these paths unsafe. A Ponzu archive must only create a sub-tree.
This may concern those who maintain package management around tar: Traditionally, package systems built around tar have used relative paths or paths of / to start the archive.

All Ponzu archives are given a prefix. This prefix could be interpeted as a suggestion – e.g. an archive with the prefix `libgizmo-1.33.7` may be overridden with simply `libgizmo` or even ignored should the implementation decide to do so. Should an implementation wish, it could override the prefix with no or little ill effect.

Not described here is verifying archive authenticity or provenance. A compliant implementation may add additional records for such things as digital signatures. Additional, implementation-dependent keys may be added to the Start of Archive record to add a digital signature for the complete archive, for instance. This is not covered in version 1 of this specification.

## Checksums

All checksums in version 1 of Ponzu are BLAKE2b-512 as defined by [RFC 7693](https://www.rfc-editor.org/rfc/rfc7693).

The preamble contains two checksums:

- The checksum of the metadata portion
- The checksum of the body content *after* compression

If there is no relevant content, the checksum must either be all zeroes (valid, but discouraged) or the null hash. For Blake2b-512, this value should be ``786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce`` in compliant implementations. This value can be computed and verified with the following Go program:

```go
package main
import (
	"fmt"
	"golang.org/x/crypto/blake2b"
)

func main() {
	h := blake2b.Sum512([]byte{})
	fmt.Printf("%x", h)
}
```

Implementations are free to determine how they present errors in validation, but must include a mechanism to be informed about a failure in data validation. 

# Appendix: Structures for Metadata maps

This section describes the metadata mapping used for each operating system.

All metadata entries are optional.

## Common

| Key         | index | type      | Since | Description                                                            |
| ----------- | ----- | --------- | ----- | ---------------------------------------------------------------------- |
| createdTime | -     | timestamp | 1     | the creation time of the file                                          |
| fileSize    | -     | uint64    | 1     | The final size on disk of the file, after reassembly and decompression |
| mimetype    | -     | string    | 1     | If applicable, the MIME filetype                                       |
| comment     | -     | string    | 1     | A freeform string comment                                              |

## UNIX

This encompasses most UNIX-like operating systems.

| Name  | index | Since | type                    | Description                         |
| ----- | ----- | ----- | ----------------------- | ----------------------------------- |
| owner | 0     | 1     | string                  | Owning user                         |
| group | 1     | 1     | string                  | Owning Group                        |
| mode  | 2     | 1     | uint16                  | File permissions (chmod compatible) |
| attr  | -     | 1     | array of string         | Attributes/flags                    |
| xattr | -     | 1     | map of string to binary | Extended Attributes                 |

## Linux

The Linux metadata contains the numbered UNIX metadata as well as the following:

| Name            | index | Since | type   | Description            |
| --------------- | ----- | ----- | ------ | ---------------------- |
| selinux_label   | -     | 1     | string | SELinux label          |
| selinux_context | -     | 1     | string | SELinux Context        |
| caps            | -     | 1     | uint64 | Linux capability flags |

## POSIX

The POSIX environment contains the numbered UNIX metadata as well as

| Name | index | since | type            | Description                                   |
| ---- | ----- | ----- | --------------- | --------------------------------------------- |
| acls | -     | 1     | Array of string | POSIX ACLs in the format described by setfacl |

The POSIX ACLs are here for historical completeness.

## WinNT

| Name       | index | type   | Description                  |
| ---------- | ----- | ------ | ---------------------------- |
| sddlString | 0     | string | SDDL ACL for the file        |
| attributes | 1     | uint16 | Windows NTFS attribute flags |

## MacOS/Darwin

The MacOS/Darwin metadata is inherited from the UNIX/BSD metadata.

# Appendix: License

This text is licensed under a Creative Commons CC BY-SA 4.0 license. For more information see https://creativecommons.org/licenses/by-sa/4.0/

In short:

You may adapt and share that adaptation of this standard with others, so long as you provide attribution and your modifications are shared under the same license.

As the Creative Commons license is not easily applicable to code, the reference implementations are under a suitable license.