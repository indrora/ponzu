package format

import (
	"github.com/indrora/ponzu/ponzu/format/metadata"
)

// There are a handful of record types. These are CBOR types.

/*
*
Record base that all other things are based on.
This base structure contains the relevant information that goes into or comes out of
the preamble, plus a Reader that is used in context.
*/

type RecordBase struct {
	/* for the future ;) */
	Metadata metadata.RecordMetadata `cbor:"osMetadata"`
}

// All archives start with a Start of Archive header
type StartOfArchive struct {
	RecordBase
	// Version of the archive
	Version uint8 `cbor:"version"`
	// Host OS the archive was made on
	Host string `cbor:"host"`
	// Prefix to write all files to
	Prefix string `cbor:"prefix"`
	// Comment for the archive (open text field)
	Comment string `cbor:"comment"`
}

type File struct {
	RecordBase
	Name string `cbor:"name"`
}

type Link struct {
	Name   string `cbor:"name"`
	Target string `cbor:"target"`
}

type Symlink struct {
	RecordBase
	Link
}
type Hardlink struct {
	RecordBase
	Link
}
type Directory struct {
	RecordBase
	Name string `cbor:"name"`
}
type ZstdDictionary struct{ RecordBase }

type OSSpecial struct {
	RecordBase
	Name        string `cbor:"name"`
	SpecialType string `cbor:"type"`
	Mode        uint32 `cbor:"mknodMode"`
	Device      uint32 `cbor:"mknodDev"`
}

type UnknownType map[string]any

type RecordInfo struct {
	StartOfArchive
	File
	Symlink
	Hardlink
	Directory
	ZstdDictionary
	OSSpecial

	// fallback
	UnknownType
}
