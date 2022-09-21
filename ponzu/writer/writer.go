package writer

import (
	"io"
	"io/fs"

	"github.com/fxamacker/cbor/v2"
	"github.com/indrora/ponzu/ponzu/format"
	pio "github.com/indrora/ponzu/ponzu/ioutil"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

var (
	ErrMisalignedWrite = errors.New("unexpected number of bytes written")
)

type WCSeeker interface {
	io.Writer
	io.Closer
	io.Seeker
}
type ArchiveWriter struct {
	fileio  WCSeeker
	blockio pio.BlockWriter
	cHeader format.StartOfArchive
}

func NewWriter(file WCSeeker, soa format.StartOfArchive) *ArchiveWriter {
	return &ArchiveWriter{fileio: file, blockio: *pio.NewBlockWriter(file, format.BLOCK_SIZE), cHeader: soa}
}

func (archive *ArchiveWriter) AppendSOA(prefix string, comment string) error {
	// write the initial header to the file.

	// This is the CBOR portion.
	archiveHeader := format.StartOfArchive{
		Version: format.PONZU_VERSION,
		Host:    format.HOST_OS_GENERIC,
		Prefix:  prefix,
		Comment: comment,
	}

	cborHeader, err := cbor.Marshal(archiveHeader)
	if err != nil {
		return errors.Wrap(err, "Failed to marshal header")
	}

	return nil
}

func (archive *ArchiveWriter) appendShort(rtype format.RecordType, flags format.RecordFlags, body []byte) error {

	// Align everything
	archive.blockio.Align()
	// Get the preamble bytes

	preamble := format.NewPreamble(rtype, flags, length)

	if int64(len(body)) > format.BLOCK_SIZE-int64(len(preamble.ToBytes())) {
		return errors.New("too much CBOR data")
	}

	preamble.Checksum = sha3.Sum512(body)

}

func (archive *ArchiveWriter) AppendStream(path string, info fs.FileInfo, stream io.Reader) error {
	// write the file to the end of the archive.
	return nil

}

func (archive *ArchiveWriter) AppendFile(path string, source string, compression format.CompressionType) error {
	return nil
}

func (archive *ArchiveWriter) AppendDirectory(path string, info fs.FileInfo) error {
	return nil
}

func (archive *ArchiveWriter) AppendSymlink(path string, destination string, info fs.FileInfo) error {
	return nil
}
