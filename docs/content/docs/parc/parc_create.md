---
weight: 100
title: "Create archives"
description: "Create a Ponzu archive"
icon: "article"
date: "2023-12-19T23:21:03-08:00"
lastmod: "2023-12-19T23:21:03-08:00"
draft: false
toc: true
---


## parc create

Create a Ponzu archive

### Synopsis

Create an archive from a specified series of glob patterns

Example globbing patterns:

* foo/ (Selects the directory "foo" but no contents)
* foo/* (Selects all contents of "foo")
* foo/** (Selects all contents of "foo" recursively)
* foo/*.txt (Selects all ".txt" files in "foo")
* foo/*/*.txt (Selects all ".txt" files in subdirectories of "foo")
* foo/*.{txt,md} (Selects all ".txt" and ".md" files in "foo")
* foo/{a,b,c}/* (Selects all contents of "foo/a", "foo/b" and "foo/c" non-recursively)

Use ? to specify a single character (foo/??/* selects all contents of two-character subdirectories of "foo")

Double stars act mostly like bash's globstar: **.txt is the same as *.txt, but foo/**/*.txt selects all .txt files in any depth subdirectory of foo.

Depending on your shell, you may have to enclose globbing patterns in single quotes('foo/**').


```
parc create [flags]
```

### Examples

```
parc create myarchive.pzarc a/** foo
```

### Options

```
      --brotli                        use Brotli compression vs. ZStandard
      --buff-size uint                Number of blocks to read into memory at once (default 5000, 2GB) (default 5000)
      --chdir string                  Search this path to find relative paths (default ".")
      --comment string                Add comment to archive
  -h, --help                          help for create
      --no-compress                   Disable compression
      --prefix string                 Archive prefix
      --zstandard-dictionary string   Path to ZStandard Dictionary to use
```

### Options inherited from parent commands

```
  -v, --verbose   Write detailed information to the terminal
```

### SEE ALSO

* [parc](parc.md)	 - Parc is a reference Ponzu ARChive tool

