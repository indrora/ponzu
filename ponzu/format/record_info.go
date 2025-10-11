package format

import (
	"github.com/indrora/ponzu/ponzu/format/metadata"
)

/*
*
Record base that all other things are based on.
This base structure contains the relevant information that goes into or comes out of
the preamble, plus a Reader that is used in context.
*/

type RecordBase struct {
	Metadata *metadata.RecordMetadata `cbor:"osMetadata"`
}

// All archives start with a Start of Archive header
type StartOfArchive struct {
	*RecordBase

	Version *int    `cbor:"version,omitempty"` // Version of the archive
	Host    *string `cbor:"host,omitempty"`    // Host OS the archive was made on
	Prefix  *string `cbor:"prefix,omitempty"`  // Prefix to write all files to
	Comment *string `cbor:"comment,omitempty"` // Comment for the archive (open text field)

}

// File record: Records its name, size on disk. All other things are encoded in the metadata block.
type File struct {
	*RecordBase
	Name *string `cbor:"name,omitempty"`
}

// Shell type for a link. Hardlinks and Softlinks are otherwise identical, but for syntactic purposes in Go, this is easier.
type Link struct {
	Name   *string `cbor:"name,omitempty"`
	Target *string `cbor:"target,omitempty"`
}

// Symbolic link
type Symlink struct {
	*RecordBase
	*Link
}

// Hard link
type Hardlink struct {
	*RecordBase
	*Link
}

// A directory is like a file, but it doesn't have a dimensionality to it.
type Directory struct {
	*RecordBase
	Name *string `cbor:"name,omitempty"`
}

// ZStandard dictionaries are weird in that they are interpreted transparently
// and are not usually processed by end applications.
type ZstdDictionary struct{ *RecordBase }

// OS Special devices. They might have dimensionality but that should get added later
// for things like block devices.

type OSSpecial struct {
	RecordBase
	Name        *string `cbor:"name,omitempty"`
	SpecialType *string `cbor:"type,omitempty"`
	Mode        *uint32 `cbor:"mknodMode,omitempty"`
	Device      *uint32 `cbor:"mknodDev,omitempty"`
}

// A fallback type that turns into an arbitrary map.
type UnknownType struct {
	Info *map[string]any
}

// a RecordInfo holds the various types of record information
// it should not be directly cbor encoded.
type RecordInfo struct {
	*StartOfArchive `cbor:"-"`
	*File           `cbor:"-"`
	*Symlink        `cbor:"-"`
	*Hardlink       `cbor:"-"`
	*Directory      `cbor:"-"`
	*ZstdDictionary `cbor:"-"`
	*OSSpecial      `cbor:"-"`
	*UnknownType    `cbor:"-"`
}
