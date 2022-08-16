# Ponzu: The new standard linear archive

This repository contains the specification for Ponzu, a linear archive format, and Parc, the Ponzu ARChiver reference implementation.

For information on the format itself, see [the spec](docs/spec.md).

For information on the archiver utility, see [the docs](docs/parc.md)

# Why? 

Because tar Sucks. The full details are included in the spec rationale, but
short form is that Tar has a lot of extensions that were meant to unify the
standard but only ended up making it worse. 

Worse, different implementations of the Tar archive format have different
and bad caveats, pitfalls, etc. 

Additionally, because it has no concept of compression, Tar fails to handle
large, sparse datasets that are spread across multiple files. 

# License

The text of the Ponzu spec is given CC-BY-SA 4.0
The reference implementation is given MIT-0

