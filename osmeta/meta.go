package osmeta

import (
	"github.com/indrora/ponzu/ponzu/format/metadata"
)

// GetMetadata retrieves OS-specific metadata for the file at the given path.
// The implementation varies by platform and is selected at compile time via build tags.
// Returns a RecordMetadata struct populated with platform-appropriate fields.
func GetMetadata(path string) (*metadata.RecordMetadata, error) {
	return getMetadata(path)
}

// SetMetadata applies OS-specific metadata to the file at the given path.
// The implementation varies by platform and is selected at compile time via build tags.
// Not all metadata fields may be applicable or settable on all platforms.
func SetMetadata(path string, meta *metadata.RecordMetadata) error {
	return setMetadata(path, meta)
}

// Platform-specific implementations must provide these functions:
// - getMetadata(path string) (*metadata.RecordMetadata, error)
// - setMetadata(path string, meta *metadata.RecordMetadata) error
