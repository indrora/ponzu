---
weight: 100
title: "Introduction"
description: ""
icon: "article"
date: "2023-12-19T22:26:17-08:00"
lastmod: "2023-12-19T22:26:17-08:00"
draft: false
toc: true
---


Ponzu is a new archive format designed by Morgan `indrora` Gangwere as a replacement for Tar. 

Ponzu is an append-only, flexible archive format, focusing on making it easy to extend in a clean and effective way. Its goal is to learn from the mistakes that Tar made in the 80s, bring in features that became prominent in other archive formats over time. 

## Why replace `tar(5)`?

Reading through the tar specification is a blast:

- Two kinds of tar archives (three if you consider GNU)
- The end of the file is marked with several zero-filled records
- Extensions. `pad[12]`. `devmajor`/`devminor`.
- `../../../../../../../etc/passwd` and other tarbombs

However, what makes Tar useful is that it’s pretty much “read header, set idx=0, crank the read head”.
The constant string manipulation that has to be done as well as a quadratic extraction time
for certain GNU tar files means that there are Many ways that the format could be improved.

For an interesting look at what tar is like in practice, see

- https://mort.coffee/home/tar/
- https://invisible-island.net/autoconf/portability-tar.html
- https://superuser.com/questions/1633073/why-are-tar-xz-files-15x-smaller-when-using-pythons-tar-library-compared-to-mac
- https://mgorny.pl/articles/portability-of-tar-features.html

Today, archive formats like zip, tar, and cpio are abused to be useful for tasks other than their original intentions. Unfortunately, this means that there are myriad ways of interpreting them, some more valid than others: zip can be read from the front to the back, despite its format declaring back to front reads. Debian packages are two TAR archives butted up against one another with some mangling. Similarly, Alpine's APK binary packages are highly mangled TAR archives. 

## What tar did right

Tar (and to a similar extent cpio) did a handful of things right:

- Fixed size blocks (tar) and straightforward design (cpio)
- Append-only record formats have their place

## How do we replace `tar(5)`

Where tar chose fixed size records of 512 bytes (the standard tape record length and block size used on most hard drives), Ponzu uses 4KiB blocks. This has three main advantages:

- Records are able to handle Extremely Long filenames, 1K UTF-8 codepoints
- Modern HDDs are aligned on 4k sectors (the “4Kn” SATA standard, ca. 2010)
- Many modern processors use 4k pages in memory

Additionally, compression should be considered free in today’s age.
Modern compression algorithms (ZStandard and Brotli, in Ponzu’s case) are reaching their theoretical maximums for certain forms of data. As such, there’s no decompression tradeoff for allowing compressed segments, only a cost during compression.
Since archival is typically a one-time operation compared to the subsequent unpacking, this tradeoff is well worth it.

By considering compression a *built-in feature*, certain things can be improved upon:

- Dictionaries can be precomputed for a corpus and compression *extended*
- Varying types and qualities of compression can be applied to make efficient use of space.

By taking into consideration the source and host operating system, more information can be archived with less loss of fidelity (e.g. Windows SIDs).

## What Ponzu does not aim to do

Ponzu does not aim to be "better" in certain respects:

* It is not "space efficient" -- up to 3KB can be wasted on a single header alone. 
* It is not designed to be "Perfect" -- It is designed to be good enough for Many Uses. 
* It does not aim to compete with ZIP and other archives like it.

