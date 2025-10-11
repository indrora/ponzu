//go:build linux

package osmeta

import (
	"fmt"

	"github.com/indrora/ponzu/ponzu/format/metadata"
	"golang.org/x/sys/unix"
)

func getMetadata(path string) (*metadata.RecordMetadata, error) {
	// First, get all Unix metadata using the shared implementation
	meta, err := getMetadataUnix(path)
	if err != nil {
		return nil, err
	}

	// Add Linux-specific metadata: SELinux context
	if context, err := getSelinuxContext(path); err == nil && context != "" {
		meta.SelinuxContext = &context
	}

	// Add Linux capabilities
	if caps, err := getCapabilities(path); err == nil && caps != 0 {
		meta.Capabilities = &caps
	}

	return meta, nil
}

func setMetadata(path string, meta *metadata.RecordMetadata) error {
	// First, set all Unix metadata
	if err := setMetadataUnix(path, meta); err != nil {
		return err
	}

	// Set SELinux context if provided
	if meta.SelinuxContext != nil {
		if err := setSelinuxContext(path, *meta.SelinuxContext); err != nil {
			return fmt.Errorf("failed to set SELinux context: %w", err)
		}
	}

	// Set capabilities if provided
	if meta.Capabilities != nil {
		if err := setCapabilities(path, *meta.Capabilities); err != nil {
			return fmt.Errorf("failed to set capabilities: %w", err)
		}
	}

	return nil
}

func getSelinuxContext(path string) (string, error) {
	// Try to get SELinux context via getxattr
	// SELinux contexts are stored in the "security.selinux" extended attribute
	size, err := unix.Getxattr(path, "security.selinux", nil)
	if err != nil {
		return "", err
	}

	buf := make([]byte, size)
	_, err = unix.Getxattr(path, "security.selinux", buf)
	if err != nil {
		return "", err
	}

	// Remove null terminator if present
	if len(buf) > 0 && buf[len(buf)-1] == 0 {
		buf = buf[:len(buf)-1]
	}

	return string(buf), nil
}

func setSelinuxContext(path string, context string) error {
	// Set SELinux context via setxattr
	return unix.Setxattr(path, "security.selinux", []byte(context), 0)
}

func getCapabilities(path string) (uint64, error) {
	// Try to get file capabilities via getxattr
	// File capabilities are stored in the "security.capability" extended attribute
	// This is a binary format, but we'll try to extract the basic capability set

	size, err := unix.Getxattr(path, "security.capability", nil)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, size)
	_, err = unix.Getxattr(path, "security.capability", buf)
	if err != nil {
		return 0, err
	}

	// The capability format is complex, but for basic storage we can just
	// return a simplified representation. A proper implementation would parse
	// the VFS_CAP_REVISION_* format.
	// For now, we'll store 0 to indicate capabilities exist but aren't parsed
	if len(buf) > 0 {
		return 0, nil
	}

	return 0, fmt.Errorf("no capabilities")
}

func setCapabilities(path string, caps uint64) error {
	// Setting capabilities requires proper VFS_CAP format encoding
	// This is a placeholder - a full implementation would encode caps properly
	// For now, we'll skip setting capabilities as it requires careful handling
	return fmt.Errorf("setting capabilities not yet implemented")
}

// Note: Linux satisfies the unix build tag, so getMetadataUnix() and setMetadataUnix()
// from meta_unix.go are available to use here
