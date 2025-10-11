//go:build !unix && !windows

package osmeta

import (
	"fmt"
	"os"

	"github.com/indrora/ponzu/ponzu/format/metadata"
)

// getMetadata provides a fallback implementation for systems that are neither Unix nor Windows.
// This "universe" implementation only provides basic file metadata that's available through
// the standard library.
func getMetadata(path string) (*metadata.RecordMetadata, error) {
	meta := &metadata.RecordMetadata{}

	// Get basic file info
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Populate only common metadata available on all platforms
	modTime := fileInfo.ModTime()
	meta.ModifiedTime = &modTime
	fileSize := uint64(fileInfo.Size())
	meta.FileSize = &fileSize

	// No creation time, owner, group, or OS-specific attributes available
	// in the fallback implementation

	return meta, nil
}

// setMetadata provides a fallback implementation for setting metadata.
// Only basic operations are supported.
func setMetadata(path string, meta *metadata.RecordMetadata) error {
	// Only timestamps can be set in the fallback implementation
	if meta.ModifiedTime != nil {
		atime := *meta.ModifiedTime
		mtime := *meta.ModifiedTime

		if err := os.Chtimes(path, atime, mtime); err != nil {
			return fmt.Errorf("failed to set times: %w", err)
		}
	}

	// Other metadata cannot be set on unknown platforms
	// Silently ignore rather than error

	return nil
}
