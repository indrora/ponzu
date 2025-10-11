//go:build unix && !(aix || ppc64)

package osmeta

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func getXattrs(path string) (map[string][]byte, error) {
	// List all extended attribute names
	size, err := unix.Listxattr(path, nil)
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return nil, nil
	}

	buf := make([]byte, size)
	size, err = unix.Listxattr(path, buf)
	if err != nil {
		return nil, err
	}

	xattrs := make(map[string][]byte)

	// Parse null-terminated attribute names
	start := 0
	for i := 0; i < size; i++ {
		if buf[i] == 0 {
			name := string(buf[start:i])

			// Get the value for this attribute
			valueSize, err := unix.Getxattr(path, name, nil)
			if err != nil {
				continue // Skip attributes we can't read
			}

			value := make([]byte, valueSize)
			_, err = unix.Getxattr(path, name, value)
			if err != nil {
				continue
			}

			xattrs[name] = value
			start = i + 1
		}
	}

	return xattrs, nil
}

func setXattrs(path string, xattrs map[string][]byte) error {
	for name, value := range xattrs {
		if err := unix.Setxattr(path, name, value, 0); err != nil {
			return fmt.Errorf("failed to set xattr %s: %w", name, err)
		}
	}
	return nil
}
