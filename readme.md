# Ponzu: The new standard linear archive

This repository contains the specification for Ponzu, a linear archive format, and Parc, the Ponzu ARChiver reference implementation.

For information on the format itself, see [the spec](docs/content/docs/spec.md).

# Why? 

Because tar Sucks. The full details are included in the spec rationale, but
short form is that Tar has a lot of extensions that were meant to unify the
standard but only ended up making it worse. 

Worse, different implementations of the Tar archive format have different
and bad caveats, pitfalls, etc. 

Additionally, because it has no concept of compression, Tar fails to handle
large, sparse datasets that are spread across multiple files. 

# Building

To build the reference archiver, run

```sh
make clean
make all
```

This will compile the `parc` binary in `bin/`. 

# Testing

To run the test suite, run 

```sh
make test
```

## Spewstat

During development, I needed a mechanism to print the OS-specific stat information.
spewstat (`make spewstat`) dumps the internal Go representation of

* stat
* stat.Sys()

and Xattrs. 


# License

The text of the Ponzu spec is given CC-BY-SA 4.0
The ponzu library in Go is MIT.
The reference archiver implementation is given MIT-0. 

for more information, see [LICENSE.md](LICENSE.md)