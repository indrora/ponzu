//go:build unix && !linux && !darwin

package osmeta

import (
	"github.com/indrora/ponzu/ponzu/format/metadata"
)

// getMetadata is the Unix implementation (non-Linux Unix systems)
func getMetadata(path string) (*metadata.RecordMetadata, error) {
	return getMetadataUnix(path)
}

// setMetadata is the Unix implementation (non-Linux Unix systems)
func setMetadata(path string, meta *metadata.RecordMetadata) error {
	return setMetadataUnix(path, meta)
}
