package writer

import (
	"errors"
	"fmt"
	"io"

	"github.com/indrora/parc/pitch/format"
	"github.com/indrora/parc/pitch/ioutil"
)

var (
	UnexpectedWriteCount = errors.New("unexpected number of bytes written")
)

type ArchiveWriter struct {
	fileio   io.ReadWriteSeeker
	cHeader  format.PitchArchiveHeader
	zStdDict []byte
}

func NewWriter(file io.ReadWriteSeeker) ArchiveWriter {
	return ArchiveWriter{fileio: file}
}

func (archive *ArchiveWriter) BeginArchive(header format.PitchArchiveHeader) error {
	block := make([]byte, format.BLOCK_SIZE)

	written, err := archive.fileio.Write(block)
	if err != nil {
		return err
	}
	if int64(written) != format.BLOCK_SIZE {
		return fmt.Errorf("%w: Wrote %d bytes, wanted to write %d", UnexpectedWriteCount, written, format.BLOCK_SIZE)
	}
	return nil
}

func (archive *ArchiveWriter) Append(info format.PitchFileHeader, reader io.Reader) {

	if archive.cHeader.Compression == format.COMPRESSION_NONE {
		// Just copy the contents over
	}

}

func (archive *ArchiveWriter) getCompressionStream(reader io.Reader) (io.Reader, error) {

	switch archive.cHeader.Compression {
	case format.COMPRESSION_NONE:
		return reader, nil
	case format.COMPRESSION_ZSTD:
		return ioutil.ZstdWriter
	default:

	}

}
