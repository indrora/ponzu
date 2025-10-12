---
weight: 200
title: "The Ponzu Spec"
draft: false
---

This document outlines the specification for the Ponzu archive format. 

# Introduction

Ponzu archives are comprised of *records* aligned on 4K boundaries. 

Each record looks like this: 

```goat 
                              |<-- modulo  -->|
.--------.-------- ~ --.-- ~ -|---------------.-- ~ ----.
| header | Information |  Body data.          | padding |
'--------'-------- ~ --'-- ~ -|---------------'-- ~ ----'
.<-- padded to 4KiB --> <--    block_count x 4KiB    -->.
```

A complete archive looks like this:

```goat
.<- one-> <---------------------------- many --------------------------------> <-one ->.
.--------.---------- ~~ --------.------------- ~~ -----.------------- ~~ -----.--------.
| start  | Info   | record body | info   | record body | info   | record body | end    |
| record | + meta | data blocks | + meta | data blocks | + meta | data blocks | record |
.--------'---------- ~~ --------'------------- ~~ -----'------------- ~~ -----'--------.
```

Archives may be appended to one another. In such a case, each should be considered independent. 


# The most minimal Ponzu archive

The most minimal Ponzu archive consists purely of two control records: a `CONTROL_START` record and a `CONTROL_END` record. 

# Record structure

Each Ponzu record is headed by a Preamble, followed by a [RFC 8949 Concise Binary Object Representation (CBOR)](https://www.rfc-editor.org/rfc/rfc8949)
encoded record information block, padded with zeros to the next 4KiB boundary, then zero or more 4KiB aligned bytes of data, padded with zeros to the next 4KiB boundary.


## Preamble fields

A preamble consists of the following:

- The characters ‘PONZU\0’ - including the null termination.
- A one byte record type (uint8_t)
- A one-byte compression type
- A two-byte (uint16_t) flag field.
- A uint64_t defining the number of 4KiB data segments to follow
- A uint16_t defining the number of bytes used in the final data block
- A 64-byte (512-bit) BLAKE2b-512 checksum of the record info section
- A uint16_t defining the length of the record info section
- A 64-byte (512 bits) BLAKE2b-512 checksum of content

A C implementation of the standard might use something like this:

```c
// values are Big-Endian (“Network Order”) on disk.
__attribute__((packed))
struct RECORD_PREAMBLE {
    uint8_t  magic[6];           // "PONZU\0"
    uint8_t  record_type;        // 0 = SOA, 1 = file, etc.
    uint8_t  compression;        // Type of compression
    uint16_t flags;              // Flag Set
    uint64_t data_len;           // # of segments to read
    uint16_t data_modulo;        // # of bytes to use in last block
    uint8_t  data_checksum[64];  // BLAKE2b-512 of the data, post-compression.
    uint16_t info_length;        // length of the metadata to be read
    uint8_t  info_checksum[64];  // BLAKE2b-512 checksum of the metadata 
};
```

The defined record types are

| Type value | Introduced | Name                 | Description                                                     | Has data? |
| ---------- | ---------- | -------------------- | --------------------------------------------------------------- | --------- |
| 0          | 1          | Control              | Start, end, or other “special” actions for the archive          | No        |
| 1          | 1          | File                 | A regular file.                                                 | Varies    |
| 2          | 1          | Symlink              | A symbolic link to a path                                       | No        |
| 3          | 1          | Hardlink             | A hard link to a specific inode                                 | No        |
| 4          | 1          | Directory            | A directory                                                     | No        |
| 5          | 1          | Zstandard Dictionary | Dictionary for ZStandard to use during decompression.           | Varies    |
| 126        | 1          | OS Special           | An OS-Special inode                                             | No        |
| 127        | 1          | Continuation block   | Continuation of the previous record                             | Yes, 1+   |
| >127       | 1          | Reserved             | All values > 127 are reserved for implementation defined usage. | arbitrary |


## Flags 

The following flags are used:

| Value    | Introduced | Name               | Description                                                          | Context |
| -------- | ---------- | ------------------ | -------------------------------------------------------------------- | ------- |
| `0b0001` | 1          | `CONTROL_START`    | (for a control record) This is the start of an archive.              | Control |
| `0b0010` | 1          | `CONTROL_END`      | (For a control record) This is the end of an archive.                | Control |
| `0b0100` | 1          | `CONTROL_STREAMED` | (for a control record) This archive may not contain checksums.       | Control |
| `0b0001` | 1          | `CONTINUES`        | (For any record) This record has continuation blocks that follow it. | Any N>0 |

*Note*: The `CONTROL_START` flag is only valid in the context of a control record. `CONTINUES` is only valid in the context of
a record whose body data is >0 segments.

## Record information

All Ponzu record information is encoded as CBOR.

### Example records

The following JSON structures outline what common entries look like. These map 1:1 with their CBOR equivalents.


File:
```json
{
    "name":"example.txt",
    "osMetadata":{
        "universe.fileSize":420,
        "universe.comment":"What a lovely day we're having?",
        "winnt.sddlString":"O:AOG:DAD:(A;;RPWPCCDCLCSWRCWDWOGA;;;S-1-0-0)",
    }
}
```

Symlink:
```json
{
    "name":"localtime",
    "target":"time/pacific",
    "osMetadata": {
        "unix.mode": 416,
        "unix.owner":"root",
        "unix.group":"root",
    }
}
```

Start of archive (Control with CONTROL_START set):

```json
{
    "version":1,
    "host":"winnt",
    "prefix":"superduper",
    "comment":"v0.0.1-git.9ad4b5",
}
```

### Example encoded records

This is a start of archive record:

```
Hex View  00 01 02 03 04 05 06 07  08 09 0A 0B 0C 0D 0E 0F
 
00000000  50 4F 4E 5A 55 00 00 00  00 01 00 00 00 00 00 00  PONZU...........
00000010  00 00 00 00 78 6A 02 F7  42 01 59 03 C6 C6 FD 85  ....xj..B.Y.....
00000020  25 52 D2 72 91 2F 47 40  E1 58 47 61 8A 86 E2 17  %R.r./G@.XGa....
00000030  F7 1F 54 19 D2 5E 10 31  AF EE 58 53 13 89 64 44  ..T..^.1..XS..dD
00000040  93 4E B0 4B 90 3A 68 5B  14 48 B7 55 D5 6F 70 1A  .N.K.:h[.H.U.op.
00000050  FE 9B E2 CE 00 5A E4 5D  04 72 10 18 D9 B8 14 34  .....Z.].r.....4
00000060  D4 07 80 52 75 04 15 FC  75 3A 25 D8 7D E2 2C 2A  ...Ru...u:%.}.,*
00000070  A0 DE 29 A1 AA 49 43 72  F5 2A C7 8E 64 4B D0 B8  ..)..ICr.*..dK..
00000080  3B FC 9E 87 97 8B 38 05  A2 53 60 A5 84 54 EE 65  ;.....8..S`..T.e
00000090  B3 2C 8F D2 68 53 A5 6A  6F 73 4D 65 74 61 64 61  .,..hS.josMetada
000000A0  74 61 A1 74 75 6E 69 76  65 72 73 65 2E 63 72 65  ta.tuniverse.cre
000000B0  61 74 65 64 54 69 6D 65  1A 68 EA 0B 82 67 76 65  atedTime.h...gve
000000C0  72 73 69 6F 6E 01 64 68  6F 73 74 68 75 6E 69 76  rsion.dhosthuniv
000000D0  65 72 73 65 66 70 72 65  66 69 78 60 67 63 6F 6D  ersefprefix`gcom
000000E0  6D 65 6E 74 6B 48 65 6C  6C 6F 20 57 6F 72 6C 64  mentkHello World
000000F0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................

                / -- padded to 4KiB boundary -- /


```

This is a file record:

```
Hex View  00 01 02 03 04 05 06 07  08 09 0A 0B 0C 0D 0E 0F
 
00005000  50 4F 4E 5A 55 00 01 00  00 00 00 00 00 00 00 00  PONZU...........
00005010  00 01 00 66 00 A8 DC 16  4B 7C 62 43 A5 67 35 F2  ...f....K|bC.g5.
00005020  E5 FA 83 5B E9 2B A2 A0  C3 AE 15 64 A8 F3 3C 22  ...[.+.....d..<"
00005030  A9 04 17 EB D8 AE 9A ED  17 92 36 39 5F F1 3F 20  ..........69_.? 
00005040  63 49 3C BC 55 A4 BF 31  89 4B CD 2C E7 34 F7 D7  cI<.U..1.K.,.4..
00005050  2C 9C 74 BC 00 A8 20 A0  CB C1 73 9B 16 CA 62 BD  ,.t... ...s...b.
00005060  3B DE 45 36 F9 CB 15 07  A2 50 26 14 DA D6 5F A9  ;.E6.....P&..._.
00005070  2F 9D D1 47 5A AF 50 DE  29 91 14 37 BC DE EB E3  /..GZ.P.)..7....
00005080  05 E8 79 CF FB CC 2E BB  88 D8 71 3A 37 E3 3C 7A  ..y.......q:7.<z
00005090  3D AA 36 81 C1 35 A2 6A  6F 73 4D 65 74 61 64 61  =.6..5.josMetada
000050A0  74 61 A6 74 75 6E 69 76  65 72 73 65 2E 63 72 65  ta.tuniverse.cre
000050B0  61 74 65 64 54 69 6D 65  1A 65 82 87 00 75 75 6E  atedTime.e...uun
000050C0  69 76 65 72 73 65 2E 6D  6F 64 69 66 69 65 64 54  iverse.modifiedT
000050D0  69 6D 65 1A 65 82 87 00  71 75 6E 69 76 65 72 73  ime.e...qunivers
000050E0  65 2E 66 69 6C 65 53 69  7A 65 18 66 6A 75 6E 69  e.fileSize.fjuni
000050F0  78 2E 6F 77 6E 65 72 67  69 6E 64 72 6F 72 61 6A  x.ownergindroraj
00005100  75 6E 69 78 2E 67 72 6F  75 70 65 73 74 61 66 66  unix.groupestaff
00005110  69 75 6E 69 78 2E 6D 6F  64 65 19 01 A4 64 6E 61  iunix.mode...dna
00005120  6D 65 78 1A 73 69 74 65  2F 61 72 63 68 65 74 79  mex.site/archety
00005130  70 65 73 2F 64 65 66 61  75 6C 74 2E 6D 64 00 00  pes/default.md..
00005140  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
00005150  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................

                / -- padded to 4KiB boundary -- /

00005FC0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
00005FD0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
00005FE0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
00005FF0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
00006000  2B 2B 2B 0A 74 69 74 6C  65 20 3D 20 27 7B 7B 20  +++.title = '{{ 
00006010  72 65 70 6C 61 63 65 20  2E 46 69 6C 65 2E 43 6F  replace .File.Co
00006020  6E 74 65 6E 74 42 61 73  65 4E 61 6D 65 20 22 2D  ntentBaseName "-
00006030  22 20 22 20 22 20 7C 20  74 69 74 6C 65 20 7D 7D  " " " | title }}
00006040  27 0A 64 61 74 65 20 3D  20 7B 7B 20 2E 44 61 74  '.date = {{ .Dat
00006050  65 20 7D 7D 0A 64 72 61  66 74 20 3D 20 74 72 75  e }}.draft = tru
00006060  65 0A 2B 2B 2B 0A 00 00  00 00 00 00 00 00 00 00  e.+++...........
00006070  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................

                / -- padded to 4KiB boundary -- /

00006FE0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
00006FF0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
```

### Archive Control

An archive control record is defined by its flags:

- `CONTROL_START`: This is a start of archive record.
- `CONTROL_END`: This is the end of the archive

The Start of Archive record is used to define the parameters of an archive.

| Name      | since | type   | Description                                        |
| --------- | ----- | ------ | -------------------------------------------------- |
| `version` | 1     | Uint8  | Version of the Ponzu spec this archive conforms to |
| `host`    | 1     | string | Host OS type that this archive was created on      |
| `prefix`  | 1     | string | Prefix used by all files in this archive           |
| `comment` | 1     | string | Comment, text                                      |

{{< alert icon="" context="info" >}}
 Note: The prefix MUST NOT begin with a leading / and any compliant implementation MUST discard a leading slash unless the implementation gives a mechanism to “trust” the archive.
{{< /alert >}}

The End of Archive record is simply a marker that the end of the archive has been achieved. 

### File

| Name   | Since | type   | Description |
| ------ | ----- | ------ | ----------- |
| `name` | 1     | string | filename    |

### Symlinks and Hardlinks

Links are Files with no data section and the following fields:

| Name     | Since | type   | Description |
| -------- | ----- | ------ | ----------- |
| `name`   | 1     | string | filename    |
| `target` | 1     | string | Link target |

Hardlinks MUST refer to a file within the archive and MUST NOT begin with `/`.

### Directories

| Name   | Since | type   | Description    |
| ------ | ----- | ------ | -------------- |
| `name` | 1     | string | directory name |


### ZStandard Dictionary

a ZStandard Dictionary has no specific fields, however the following optional fields
may be included:

| Name      | Since | type   | Description                                                 |
| --------- | ----- | ------ | ----------------------------------------------------------- |
| `version` | 1     | string | Version of ZStandard that created this dictionary, if known |

ZStandard dictionaries *must not* be compressed.

When a Dictionary record is received, the old dictionary (if any) should be discarded.

### OS Special

For operating systems that support “Special” files (e.g. FIFOs, device nodes, etc),
this type is used. These files generally do not contain “data”.

| Name        | Since | type   | Description                 |
| ----------- | ----- | ------ | --------------------------- |
| `name`      | 1     | string | filename                    |
| `type`      | 1     | string | only “mknod” valid for now. |
| `mknodMode` | 1     | u32    | Mode for mknod              |
| `mknodDev`  | 1     | u32    | Dev_t value for mknod       |

### Continuation record

Continuation records are used to denote additional segments of the previous record. A record declares that
there is an additional segment to follow by setting the `CONTINUES` flag. Any record with a `CONTINUES` flag
must be followed by a continuation record.

A continuation record MUST come immediately after another record with the `CONTINUES` flag. A continuation record
that is not immediately preceded by another record with its `CONTINUES` flag set is considered invalid and must raise
some sort of error. 

A Continuation record is specifically intended for several situations:

- Filesystems where >4GB files are not allowed, but an uncompressed >4GB file must be described
- Streamed archives where integrity must be assured during transit
- Data where it is infeasible to calculate a checksum for the full body in a reasonable amount of time

Continuation records have no record information.

### Operating system metadata 

All records have an optional, but *encouraged* field:

| Name         | Since | Type | Description            |
| ------------ | ----- | ---- | ---------------------- |
| `osMetadata` | 1     | map  | OS-Specific attributes |


All metadata entries are optional.

Metadata is comprised of a series of string keys and strictly typed values. Each key is prefixed with the host that it comes
from, with all common metadata being prefixed with "universe."

A file may have the metadata from multiple hosts: If a file has a valid WinNT SDDL and a valid UNIX xattr declaration, both
are valid to have in a file. 

#### Common

| Key                     | type      | Since | Description                                                            |
| ----------------------- | --------- | ----- | ---------------------------------------------------------------------- |
| `universe.createdTime`  | timestamp | 1     | the creation time of the file                                          |
| `universe.modifiedTime` | timestamp | 1     | The modification time of the file                                      |
| `universe.fileSize`     | uint64    | 1     | The final size on disk of the file, after reassembly and decompression |
| `universe.mimetype`     | string    | 1     | If applicable, the MIME filetype                                       |
| `universe.comment`      | string    | 1     | A freeform string comment                                              |

#### UNIX

This encompasses most UNIX-like operating systems.

| Name         | Since | type                    | Description                         |
| ------------ | ----- | ----------------------- | ----------------------------------- |
| `unix.owner` | 1     | string                  | Owning user                         |
| `unix.group` | 1     | string                  | Owning Group                        |
| `unix.mode`  | 1     | uint16                  | File permissions (chmod compatible) |
| `unix.attr`  | 1     | array of string         | Attributes/flags                    |
| `unix.xattr` | 1     | map of string to binary | Extended Attributes                 |

#### Linux

The Linux metadata contains the UNIX metadata as well as the following:

| Name                    | Since | type   | Description            |
| ----------------------- | ----- | ------ | ---------------------- |
| `linux.selinux_label`   | 1     | string | SELinux label          |
| `linux.selinux_context` | 1     | string | SELinux Context        |
| `linux.caps`            | 1     | uint64 | Linux capability flags |

#### POSIX

The POSIX environment contains the numbered UNIX metadata as well as

| Name         | since | type            | Description                                   |
| ------------ | ----- | --------------- | --------------------------------------------- |
| `posix.acls` | 1     | Array of string | POSIX ACLs in the format described by setfacl |

The POSIX ACLs are here for historical completeness.

#### WinNT

| Name               | since | type   | Description                  |
| ------------------ | ----- | ------ | ---------------------------- |
| `winnt.sddlString` | 1     | string | SDDL ACL for the file        |
| `winnt.attributes` | 1     | uint16 | Windows NTFS attribute flags |

#### MacOS/Darwin

The MacOS/Darwin metadata is inherited from the UNIX/BSD metadata.

| Name              | since | type   | description    |
| ----------------- | ----- | ------ | -------------- |
| `darwin.bsdFlags` | 1     | uint64 | see chflags(2) |


# Details of implementation

This section outlines specific details about the format, such as types of compression, byte order, and a discussion on security.

## Compression

Two algorithms are defined for compression in Ponzu: ZStandard and Brotli.

| value | Since | name      | Info                             |
| ----- | ----- | --------- | -------------------------------- |
| `0`   | 1     | None      |                                  |
| `1`   | 1     | ZStandard | https://facebook.github.io/zstd/ |
| `2`   | 1     | Brotli    | https://github.com/google/brotli |

Compression is applied only to the data chunks that follow a record header. For continuation records, this means that each part
will be compressed individually; there is no requirement that a continue record be compressed or use the same compression as
the other parts within the final, reassembled content. 

## Host Operating System values

The following operating systems might show up:

- `unix` - A UNIX/BSD system
- `linux` - A typical Linux system
- `posix` - A POSIX-compliant system
- `winnt` - A Windows NT system, such as Windows 11
- `darwin` - A MacOS/Darwin system
- `universe` - A generic, know-nothing system.

Darwin, POSIX and Linux are supersets of UNIX.

### The `universe` host

The `universe` value is presented as a generic: Archives with the “Universe” machine are treated more or less like large file supporting tar archives with checksums. No file attribute metadata should be inferred or included.

## Handling archives from foreign systems and future versions.

When an implementation encounters an archive that uses an unknown or future version of the specification, a compliant archive utility SHOULD provide a mechanism to extract the foreign or unknown information alongside the data portion.

If an implementation encounters an unknown compression format or file record, it SHOULD provide a means to extract the data segment of the record AS-IS, writing the content to an unambiguous filename (e.g. `filename.ponzu_data`)

## Streamed Archives

Streamed archives are generated on the fly or in situations where seeking back through the file is not reasonable (e.g. because it is a TCP socket, TTY, etc).

Streamed archives may be comprised of precomputed file records, in which the precomputed checksum is known. In these cases, an individual file record may have a checksum,
but a checksum field with all zeros for the data is acceptable.

## Character encoding

All filenames in Ponzu are UTF-8 encoded.

## Byte Order

All values shall be Big-Endian (“Network Order”), as defined by RFC8949.

*Rationale*: This is to maintain consistency with CBOR as well as firmly define the explicit order of bytes. 

## Security

Not described here is verifying archive authenticity or provenance. A compliant implementation may add additional records for such things as digital signatures. As an example, additional, implementation-dependent keys may be added to the Start of Archive record to add a digital signature for the complete archive. This is not covered in version 1 of this specification.

### Paths

A common vulnerability in Tar and other formats is path traversal attacks. These attacks are often
the result of something similar to files named `../../../../../etc/sshd/authorized-keys` and the like.

Paths (including the archive prefix) in Ponzu archives MUST be fully resolved (containing no
`..` portions.) A leading `/` is always to be interpreted as `./` except for symbolic and hard links.  

This may concern those who maintain package management around tar:
Traditionally, package systems built around tar have used relative paths or paths of / to start the archive. In this case, it is up to the implementation to provide a declared "safe" way of handling this situation. 


### Checksums

All checksums in version 1 of Ponzu are BLAKE2b-512 as defined by [RFC 7693](https://www.rfc-editor.org/rfc/rfc7693).

The preamble contains two checksums:

- The checksum of the record information portion
- The checksum of the body content *after* compression

{{< alert >}}
To be clear: it is not required to decompress the contents of the archive to verify its integrity.
{{< /alert >}}

If there is no relevant content, the checksum must either be all zeroes (valid, but discouraged) or the null hash. For Blake2b-512, this value should be `786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce` in compliant implementations. This value can be computed and verified with the following Go program:

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

# Reference implementation notes

The reference implementation, roughly based on the go `tar` implementation, has several specific quirks:

* The reader interface silently consumes ZStandard dictionary records for the purposes of decompression
* The reader interface's `Validate(...)` call advances the reader
* There is no "Rewind" within the reader interface; file position management is up to a sufficiently complex user.
* The writer interface makes an attempt to sanitize paths

The reader/writer interface depends on `https://github.com/klauspost/compress` for compression. 

An Imhex pattern has been included to validate the structure of an archive for validation. This pattern does not include
validation of the cryptographic hashes, CBOR bodies, or compression. 

# Appendix: License

This text is licensed under a Creative Commons CC BY-SA 4.0 license. For more information see https://creativecommons.org/licenses/by-sa/4.0/

In short:

You may adapt and share that adaptation of this standard with others, so long as you provide attribution and your modifications are shared under the same license.

As the Creative Commons license is not easily applicable to code, the reference implementations are under a suitable license.