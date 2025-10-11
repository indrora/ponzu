package writer

import (
	"io/fs"

	"github.com/indrora/ponzu/osmeta"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/format/metadata"
)

func (archive *ArchiveWriter) AppendDirectory(path string, info fs.FileInfo) error {

	meta, err := osmeta.GetMetadata(info.Name())

	if err != nil {
		return err
	}

	rInfo := &format.RecordInfo{
		Directory: &format.Directory{
			Name:       metadata.MakePointer(path),
			RecordBase: &format.RecordBase{Metadata: meta},
		},
	}

	err = archive.AppendBytes(format.RECORD_TYPE_DIRECTORY, format.RECORD_FLAG_NONE, format.COMPRESSION_NONE, rInfo, nil)

	return err
}

func (archive *ArchiveWriter) AppendSymlink(path string, destination string, info fs.FileInfo) error {

	meta, err := osmeta.GetMetadata(info.Name())

	if err != nil {
		return err
	}

	rInfo := &format.RecordInfo{
		Symlink: &format.Symlink{
			Link: &format.Link{
				Name:   &path,
				Target: &destination,
			},
			RecordBase: &format.RecordBase{Metadata: meta},
		},
	}

	err = archive.AppendBytes(format.RECORD_TYPE_DIRECTORY, format.RECORD_FLAG_NONE, format.COMPRESSION_NONE, rInfo, nil)

	return err

}
