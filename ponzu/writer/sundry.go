package writer

import (
	"io/fs"

	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/format/metadata"
)

func (archive *ArchiveWriter) AppendDirectory(path string, info fs.FileInfo) error {

	err := archive.AppendBytes(format.RECORD_TYPE_DIRECTORY, format.RECORD_FLAG_NONE, format.COMPRESSION_NONE, format.Directory{
		Name: path,

		RecordBase: format.RecordBase{
			Metadata: metadata.RecordMetadata{
				CommonMetadata: metadata.CommonMetadata{
					ModifiedTime: metadata.MakePointer(info.ModTime()),
				},
			},
		},
	}, nil)

	return err
}

func (archive *ArchiveWriter) AppendSymlink(path string, destination string, info fs.FileInfo) error {
	err := archive.AppendBytes(format.RECORD_TYPE_DIRECTORY, format.RECORD_FLAG_NONE, format.COMPRESSION_NONE, format.Symlink{
		Link: format.Link{
			Name:   path,
			Target: destination,
		},
		RecordBase: format.RecordBase{
			Metadata: metadata.RecordMetadata{
				CommonMetadata: metadata.CommonMetadata{
					ModifiedTime: metadata.MakePointer(info.ModTime()),
				},
			},
		},
	}, nil)

	return err

}
