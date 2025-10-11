//go:build darwin

package osmeta

import (
	"fmt"
	"syscall"

	"github.com/indrora/ponzu/ponzu/format/metadata"
)

func getMetadata(path string) (*metadata.RecordMetadata, error) {
	// First, get all Unix metadata using the shared implementation
	meta, err := getMetadataUnix(path)
	if err != nil {
		return nil, err
	}

	// Add Darwin-specific metadata: BSD flags
	if flags, err := getBsdFlags(path); err == nil && flags != 0 {
		meta.BsdFlags = &flags
	}

	return meta, nil
}

func setMetadata(path string, meta *metadata.RecordMetadata) error {
	// First, set all Unix metadata
	if err := setMetadataUnix(path, meta); err != nil {
		return err
	}

	// Set BSD flags if provided
	if meta.BsdFlags != nil {
		if err := setBsdFlags(path, *meta.BsdFlags); err != nil {
			return fmt.Errorf("failed to set BSD flags: %w", err)
		}
	}

	return nil
}

func getBsdFlags(path string) (uint64, error) {
	// Use syscall to get file flags (chflags equivalent)
	var stat syscall.Stat_t
	if err := syscall.Lstat(path, &stat); err != nil {
		return 0, err
	}

	// On Darwin, stat.Flags contains the BSD flags
	return uint64(stat.Flags), nil
}

func setBsdFlags(path string, flags uint64) error {
	// Use syscall.Chflags to set BSD flags
	return syscall.Chflags(path, int(flags))
}
