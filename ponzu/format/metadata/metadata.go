package metadata

import (
	"time"
)

func MakePointer[T any](x T) *T {
	return &x
}

// The common metadata that is used by all forms
// By all technical means, there is no "required" metadata.
type CommonMetadata struct {
	CreatedTime *time.Time `cbor:"createdTime,omitempty"`
	FileSize    *uint64    `cbor:"fileSize,omitempty"`
	MimeType    *string    `cbor:"mimetype,omitempty"`
	Comment     *string    `cbor:"comment,omitempty"`
}

// UNIX style metadata: Owner, Group, Mode, and some additional flags.
type UNIXMetadata struct {
	CommonMetadata
	Owner    *string            `cbor:"0,keyasint,omitempty"`
	Group    *string            `cbor:"1,keyasint,omitempty"`
	Mode     *uint16            `cbor:"2,keyasint,omitempty"`
	Attribs  *[]string          `cbor:"attr,omitempty"`
	Xattribs *map[string][]byte `cbor:"xattr,omitempty"`
}

// Linux metadata: SELinux additions and capability flags.
type LinuxMetadata struct {
	UNIXMetadata
	SelinuxLabel   *string `cbor:"selinux_label,omitempty"`
	SelinuxContext *string `cbor:"selinux_context.omitempty"`
	Capabilities   *uint64 `cbor:"caps,omitempty"`
}

// POSIX metadata: UNIXy, but with the additonal list of ACLs
type POSIXMetadata struct {
	UNIXMetadata
	Acls *[]string `cbor:"acls,omitempty"`
}

// WinNT metadata: NT has no concept of owning users/modes, instead places ACLs on files based on common groupings and such.
// Files have a bitfield of various attributes, as well.
type WinNTMetadata struct {
	CommonMetadata
	SddlString *string `cbor:"0,omitempty"`
	Attributes *uint16 `cbor:"1,omitempty"`
}

// MacOS/Darwin metadata: Just UNIX, for now.
type DarwinMetadata struct {
	UNIXMetadata
}

func GetMetadataForPath(filepath string) (any, error) {
	return CommonMetadata{}, nil
}
