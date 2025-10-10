---
title: "PARC reference archiver documentation"
layout: "manpage"
docfile: "doc"
---

# Building

To build the `parc` utility:

```sh
make clean
make parc
```
The compiled binary will be available in the `bin/` directory.

To generate the documentation, run 

```sh
make docs
```


# Testing

Run the test suite with:

```sh
make test
```

