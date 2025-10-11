package cmd

import (
	"fmt"

	"github.com/indrora/ponzu/ponzu/format"
)

var rtypetostr = map[format.RecordType]string{
	format.RECORD_TYPE_CONTROL:     "Control",
	format.RECORD_TYPE_FILE:        "File",
	format.RECORD_TYPE_DIRECTORY:   "Directory",
	format.RECORD_TYPE_HARDLINK:    "Hardlink",
	format.RECORD_TYPE_SYMLINK:     "Symlink",
	format.RECORD_TYPE_OS_SPECIAL:  "OS Special",
	format.RECORD_TYPE_ZDICTIONARY: "Zstd Dictionary",
}

func RecordTypeToString(r format.RecordType) string {
	if v, ok := rtypetostr[r]; ok {
		return v
	} else {
		return fmt.Sprintf("%v", r)
	}
}

func RecordInfoToString(nfo format.RecordInfo) string {
	if nfo.File != nil {
		return fmt.Sprintf("%s <created %v modified %v size %v>",
			*nfo.File.Name,
			*nfo.File.RecordBase.Metadata.CommonMetadata.CreatedTime,
			*nfo.File.RecordBase.Metadata.CommonMetadata.ModifiedTime,
			*nfo.File.RecordBase.Metadata.CommonMetadata.FileSize,
		)
	} else if nfo.Directory != nil {
		return fmt.Sprintf("%v <created %v modified  %v>",
			*nfo.Directory.Name,
			*nfo.Directory.Metadata.CommonMetadata.CreatedTime,
			*nfo.Directory.Metadata.CommonMetadata.ModifiedTime)
	} else if nfo.StartOfArchive != nil {
		return fmt.Sprintf("Archive start, prefix \"%v\" <version %v, comment \"%v\", created %v, host %v>",
			*nfo.StartOfArchive.Prefix,
			*nfo.StartOfArchive.Version,
			*nfo.StartOfArchive.Comment,
			*nfo.StartOfArchive.Metadata.CommonMetadata.CreatedTime,
			*nfo.StartOfArchive.Host,
		)
	} else if nfo.Symlink != nil {
		return fmt.Sprintf("Symlink %v -> %v",
			*nfo.Symlink.Link.Name,
			*nfo.Symlink.Link.Target,
		)
	} else if nfo.Hardlink != nil {
		return fmt.Sprintf("Hard link %v -> %v",
			*nfo.Hardlink.Link.Name,
			*nfo.Hardlink.Link.Target,
		)
	} else if nfo.ZstdDictionary != nil {
		return "ZStd dictionary"
	} else if nfo.OSSpecial != nil {
		return fmt.Sprintf("%v <OS Specialtype %v>", *nfo.OSSpecial.Name, *nfo.OSSpecial.SpecialType)
	}

	return "(Some record...?)"
}
