//go:build unix

package osmeta

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/indrora/ponzu/ponzu/format/metadata"
)

// getMetadataUnix is the shared Unix metadata getter used by all Unix-like systems
func getMetadataUnix(path string) (*metadata.RecordMetadata, error) {
	meta := &metadata.RecordMetadata{}

	// Get basic file info
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Populate common metadata
	modTime := fileInfo.ModTime()
	meta.ModifiedTime = &modTime
	fileSize := uint64(fileInfo.Size())
	meta.FileSize = &fileSize

	// Get system-specific stat info
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return meta, nil // Return partial metadata if we can't get syscall info
	}

	meta.CreatedTime = getCreationTime(stat)

	// Get owner username
	if u, err := user.LookupId(strconv.Itoa(int(stat.Uid))); err == nil {
		meta.Owner = &u.Username
	}

	// Get group name
	if g, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid))); err == nil {
		meta.Group = &g.Name
	}

	// Get file mode (permissions)
	mode := uint16(fileInfo.Mode().Perm())
	meta.Mode = &mode

	// Get extended attributes (xattr)
	xattrs, err := getXattrs(path)
	if err == nil && len(xattrs) > 0 {
		meta.Xattribs = &xattrs
	}

	return meta, nil
}

// setMetadataUnix is the shared Unix metadata setter used by all Unix-like systems
func setMetadataUnix(path string, meta *metadata.RecordMetadata) error {
	// Set timestamps if provided
	if meta.ModifiedTime != nil || meta.CreatedTime != nil {
		atime := time.Now()
		mtime := time.Now()

		if meta.ModifiedTime != nil {
			mtime = *meta.ModifiedTime
			atime = *meta.ModifiedTime // Use mtime for atime if no specific atime
		}

		if err := os.Chtimes(path, atime, mtime); err != nil {
			return fmt.Errorf("failed to set times: %w", err)
		}
	}

	// Set ownership if provided
	if meta.Owner != nil || meta.Group != nil {
		uid := -1
		gid := -1

		if meta.Owner != nil {
			if u, err := user.Lookup(*meta.Owner); err == nil {
				if uidInt, err := strconv.Atoi(u.Uid); err == nil {
					uid = uidInt
				}
			}
		}

		if meta.Group != nil {
			if g, err := user.LookupGroup(*meta.Group); err == nil {
				if gidInt, err := strconv.Atoi(g.Gid); err == nil {
					gid = gidInt
				}
			}
		}

		if uid != -1 || gid != -1 {
			if err := os.Lchown(path, uid, gid); err != nil {
				return fmt.Errorf("failed to set ownership: %w", err)
			}
		}
	}

	// Set mode if provided
	if meta.Mode != nil {
		if err := os.Chmod(path, os.FileMode(*meta.Mode)); err != nil {
			return fmt.Errorf("failed to set mode: %w", err)
		}
	}

	// Set extended attributes if provided
	if meta.Xattribs != nil {
		if err := setXattrs(path, *meta.Xattribs); err != nil {
			return fmt.Errorf("failed to set xattrs: %w", err)
		}
	}

	return nil
}
