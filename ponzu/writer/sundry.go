package writer

import (
	"io/fs"

	"github.com/indrora/ponzu/ponzu/format"
)

func (archive *ArchiveWriter) AppendDirectory(path string, info fs.FileInfo) error {

	err := archive.AppendBytes(format.RECORD_TYPE_DIRECTORY, format.RECORD_FLAG_NONE, format.COMPRESSION_NONE, format.Directory{
		File: format.File{Name: path,
			ModTime:  info.ModTime(),
			Metadata: map[string]any{},
		},
	}, nil)

	return err
}

func (archive *ArchiveWriter) AppendSymlink(path string, destination string, info fs.FileInfo) error {
	err := archive.AppendBytes(format.RECORD_TYPE_DIRECTORY, format.RECORD_FLAG_NONE, format.COMPRESSION_NONE, format.Symlink{
		Link: format.Link{
			File: format.File{Name: path,
				ModTime:  info.ModTime(),
				Metadata: map[string]any{},
			},
			Target: destination,
		},
	}, nil)

	return err

}
