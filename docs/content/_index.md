---
Title: "Ponzu: The New Standard Linear Archive"
draft: false
---

# Ponzu Archive Format

Ponzu is a modern linear archive format designed as a replacement for tar, created by Morgan "indrora" Gangwere. It addresses the fundamental issues and limitations that have plagued tar for decades while incorporating modern compression and security features.

## Why Replace Tar?

Tar has accumulated significant technical debt over the decades:

- **Multiple incompatible formats**: Two main kinds of tar archives (three with GNU extensions)
- **Complex specifications**: Quadratic extraction times, extensive string manipulation requirements
- **Security vulnerabilities**: Path traversal attacks (`../../../etc/passwd` tarbombs)
- **No native compression**: Relies on external compression, failing with large sparse datasets
- **Poor portability**: Different implementations have varying caveats and pitfalls

More information is available in the [introduction]({{< ref intro >}}). 

## What Makes Ponzu Better

### Modern Design Principles

- **4KiB blocks**: Aligned with modern hardware (4Kn SATA drives, processor memory pages)
- **Built-in compression**: Native ZStandard and Brotli support with dictionary precomputation
- **Security by design**: Path sanitization prevents traversal attacks, BLAKE2b-512 checksums ensure integrity
- **Extensible metadata**: OS-specific attributes preserved with CBOR encoding
- **Append-only format**: Simple, reliable structure that's easy to implement

### Technical Specifications

Each Ponzu record consists of:
- **Fixed preamble**: Magic bytes, record type, compression info, checksums
- **CBOR metadata**: Flexible, extensible information storage
- **Data blocks**: Compressed content in 4KiB chunks
- **Integrity verification**: BLAKE2b-512 checksums for both metadata and content

### Supported Record Types

- **Files**: Regular files with full metadata preservation
- **Directories**: Directory structure and permissions
- **Links**: Both symbolic and hard links
- **OS Special**: Device nodes, FIFOs, and other special files
- **Control**: Archive start/end markers and parameters
- **Compression dictionaries**: For enhanced compression efficiency

## Key Features

- **Cross-platform**: Supports UNIX, Linux, POSIX, Windows NT, and macOS metadata
- **Streaming capable**: Designed for both seekable and streaming scenarios
- **Future-proof**: Extensible design accommodates unknown record types gracefully
- **UTF-8 filenames**: Native Unicode support for international file names
- **Large file support**: Handles files >4GB efficiently; Theoretically infinite files can be archived.

## Implementation

The reference implementation (`parc`) is written in Go and provides a complete archiver with:
- Command-line interface compatible with common archive operations
- Library interface for integration into other applications
- Comprehensive test suite ensuring format compliance

Build the reference archiver:
```sh
make clean
make all
```

Run tests:
```sh
make test
```

## License

- **Specification**: CC-BY-SA 4.0
- **Go library**: MIT
- **Reference implementation**: MIT-0

Ponzu represents a clean slate approach to archival, learning from tar's mistakes while embracing modern computing realities.
