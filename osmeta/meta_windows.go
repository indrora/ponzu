//go:build windows

package osmeta

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/indrora/ponzu/ponzu/format/metadata"
	"golang.org/x/sys/windows"
)

func getMetadata(path string) (*metadata.RecordMetadata, error) {
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

	// Get Windows-specific attributes
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return meta, nil // Return partial metadata
	}

	// Get file attributes
	attrs, err := windows.GetFileAttributes(pathPtr)
	if err == nil {
		attrsUint16 := uint16(attrs)
		meta.Attributes = &attrsUint16
	}

	// Get creation time (Windows has explicit creation time)
	var data windows.Win32finddata
	handle, err := windows.FindFirstFile(pathPtr, &data)
	if err == nil {
		windows.FindClose(handle)
		createdTime := time.Unix(0, data.CreationTime.Nanoseconds())
		meta.CreatedTime = &createdTime
	}

	// Get SDDL (Security Descriptor Definition Language) string
	sddl, err := getSecurityDescriptor(path)
	if err == nil && sddl != "" {
		meta.SddlString = &sddl
	}

	return meta, nil
}

func setMetadata(path string, meta *metadata.RecordMetadata) error {
	// Set timestamps if provided
	if meta.ModifiedTime != nil || meta.CreatedTime != nil {
		// Windows allows setting creation time, modification time, and access time separately
		// We'll use SetFileTime for precise control

		pathPtr, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			return fmt.Errorf("failed to convert path: %w", err)
		}

		handle, err := windows.CreateFile(
			pathPtr,
			windows.FILE_WRITE_ATTRIBUTES,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
			nil,
			windows.OPEN_EXISTING,
			windows.FILE_FLAG_BACKUP_SEMANTICS,
			0,
		)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer windows.CloseHandle(handle)

		var ctime, atime, mtime *windows.Filetime

		if meta.CreatedTime != nil {
			ft := windows.NsecToFiletime(meta.CreatedTime.UnixNano())
			ctime = &ft
		}

		if meta.ModifiedTime != nil {
			ft := windows.NsecToFiletime(meta.ModifiedTime.UnixNano())
			mtime = &ft
			// Also set access time to match modification time
			atime = &ft
		}

		if err := windows.SetFileTime(handle, ctime, atime, mtime); err != nil {
			return fmt.Errorf("failed to set file time: %w", err)
		}
	}

	// Set file attributes if provided
	if meta.Attributes != nil {
		pathPtr, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			return fmt.Errorf("failed to convert path: %w", err)
		}

		if err := windows.SetFileAttributes(pathPtr, uint32(*meta.Attributes)); err != nil {
			return fmt.Errorf("failed to set attributes: %w", err)
		}
	}

	// Set SDDL if provided
	if meta.SddlString != nil {
		if err := setSecurityDescriptor(path, *meta.SddlString); err != nil {
			return fmt.Errorf("failed to set SDDL: %w", err)
		}
	}

	return nil
}

func getSecurityDescriptor(path string) (string, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	// Get the security descriptor
	var sd *windows.SECURITY_DESCRIPTOR
	err = windows.GetNamedSecurityInfo(
		pathPtr,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|
			windows.GROUP_SECURITY_INFORMATION|
			windows.DACL_SECURITY_INFORMATION|
			windows.SACL_SECURITY_INFORMATION,
		nil, nil, nil, nil,
		&sd,
	)
	if err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(sd)))

	// Convert security descriptor to SDDL string
	var sddlPtr *uint16
	err = windows.ConvertSecurityDescriptorToStringSecurityDescriptor(
		sd,
		windows.SDDL_REVISION_1,
		windows.OWNER_SECURITY_INFORMATION|
			windows.GROUP_SECURITY_INFORMATION|
			windows.DACL_SECURITY_INFORMATION|
			windows.SACL_SECURITY_INFORMATION,
		&sddlPtr,
		nil,
	)
	if err != nil {
		return "", err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(sddlPtr)))

	return syscall.UTF16ToString((*[1 << 29]uint16)(unsafe.Pointer(sddlPtr))[:]), nil
}

func setSecurityDescriptor(path string, sddl string) error {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	sddlPtr, err := syscall.UTF16PtrFromString(sddl)
	if err != nil {
		return err
	}

	// Convert SDDL string to security descriptor
	var sd *windows.SECURITY_DESCRIPTOR
	err = windows.ConvertStringSecurityDescriptorToSecurityDescriptor(
		sddlPtr,
		windows.SDDL_REVISION_1,
		&sd,
		nil,
	)
	if err != nil {
		return err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(sd)))

	// Set the security descriptor
	return windows.SetNamedSecurityInfo(
		pathPtr,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|
			windows.GROUP_SECURITY_INFORMATION|
			windows.DACL_SECURITY_INFORMATION|
			windows.SACL_SECURITY_INFORMATION,
		nil, nil, nil, nil,
	)
}
