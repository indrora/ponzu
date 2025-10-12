package cmd

import "github.com/indrora/ponzu/ponzu/format"

func GetFileMode(m *format.RecordInfo, fallback uint32) uint32 {
	// Check to see if we have UNIX modes

	if m.File != nil && m.File.Metadata.Mode != nil {
		return uint32(*m.File.Metadata.Mode)
	} else if m.Directory != nil && m.Directory.Metadata.Mode != nil {
		return uint32(*m.Directory.Metadata.Mode)
	} else if m.Symlink != nil && m.Symlink.Metadata.UNIXMetadata.Mode != nil {
		return uint32(*m.Symlink.Metadata.Mode)
	} else if m.Hardlink != nil && m.Hardlink.Metadata.Mode != nil {
		return uint32(*m.Hardlink.Metadata.Mode)
	} else if m.OSSpecial != nil && m.OSSpecial.Metadata.Mode != nil {
		return uint32(*m.OSSpecial.Metadata.Mode)
	}

	return fallback
}
