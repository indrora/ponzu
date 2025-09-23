package metadata

import (
	"time"

	"github.com/fxamacker/cbor/v2"
)

func MakePointer[T any](x T) *T {
	return &x
}

func TransmogrifyCbor[T any](meta map[any]any) (*T, bool) {
	ret := new(T)
	bdata, err := cbor.Marshal(meta)
	if err != nil {
		return nil, false
	}
	err = cbor.Unmarshal(bdata, ret)
	if err != nil {
		return nil, false
	}
	return ret, true
}

// The common metadata that is used by all forms
// By all technical means, there is no "required" metadata.
type CommonMetadata struct {
	CreatedTime  *time.Time `cbor:"universe.createdTime,omitempty"`
	ModifiedTime *time.Time `cbor:"universe.modifiedTime,omitempty"`
	FileSize     *uint64    `cbor:"universe.fileSize,omitempty"`
	MimeType     *string    `cbor:"universe.mimetype,omitempty"`
	Comment      *string    `cbor:"universe.comment,omitempty"`
}

// UNIX style metadata: Owner, Group, Mode, and some additional flags.
type UNIXMetadata struct {
	Owner    *string            `cbor:"unix.owner,omitempty"`
	Group    *string            `cbor:"unix.group,omitempty"`
	Mode     *uint16            `cbor:"unix.mode,omitempty"`
	Attribs  *[]string          `cbor:"unix.attr,omitempty"`
	Xattribs *map[string][]byte `cbor:"unix.xattr,omitempty"`
}

// Linux metadata: SELinux additions and capability flags.
type LinuxMetadata struct {
	SelinuxLabel   *string `cbor:"linux.selinux_label,omitempty"`
	SelinuxContext *string `cbor:"linux.selinux_context.omitempty"`
	Capabilities   *uint64 `cbor:"linux.caps,omitempty"`
}

// POSIX metadata: UNIXy, but with the additonal list of ACLs
type POSIXMetadata struct {
	Acls *[]string `cbor:"posix.acls,omitempty"`
}

// WinNT metadata: NT has no concept of owning users/modes, instead places ACLs on files based on common groupings and such.
// Files have a bitfield of various attributes, as well.
type WinNTMetadata struct {
	SddlString *string `cbor:"winnt.sddlString,omitempty"`
	Attributes *uint16 `cbor:"winnt.attributes,omitempty"`
}

// MacOS/Darwin metadata: Just UNIX, for now.
type DarwinMetadata struct {
	BsdFlags *uint64 `cbor:darwin.bsd_flags,omitempty`
}

type RecordMetadata struct {
	CommonMetadata
	UNIXMetadata
	LinuxMetadata
	POSIXMetadata
	WinNTMetadata
	DarwinMetadata
}

func GetMetadataForPath(filepath string) (any, error) {
	return RecordMetadata{}, nil
}
