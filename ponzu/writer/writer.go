package writer

import (
	"io"
	"io/fs"

	"github.com/fxamacker/cbor/v2"
	"github.com/indrora/ponzu/ponzu/format"
	pio "github.com/indrora/ponzu/ponzu/ioutil"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

var (
	ErrMisalignedWrite = errors.New("unexpected number of bytes written")
)

type ArchiveWriter struct {
	fileio        io.Writer
	blockio       pio.BlockWriter
	cHeader       *format.StartOfArchive
	MaxReadBuffer uint64
}

func NewWriter(file io.Writer, readBufferSize uint64) *ArchiveWriter {
	return &ArchiveWriter{
		fileio:        file,
		blockio:       *pio.NewBlockWriter(file, format.BLOCK_SIZE),
		cHeader:       nil,
		MaxReadBuffer: readBufferSize,
	}
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

	return archive.appendShort(format.RECORD_TYPE_START, format.RECORD_FLAG_NONE, cborHeader)
}

func (archive *ArchiveWriter) appendShort(rtype format.RecordType, flags format.RecordFlags, body []byte) error {

	// Align everything
	archive.blockio.Align()
	// Get the preamble bytes

	preamble := format.NewPreamble(rtype, flags, 0)

	if int64(len(body)) > format.BLOCK_SIZE-int64(len(preamble.ToBytes())) {
		return errors.New("too much CBOR data")
	}

	preamble.Checksum = blake2b.Sum512(body)

	_, err := archive.blockio.Write(preamble.ToBytes())
	if err != nil {
		return errors.Wrap(err, "failed writing preamble")
	}
	_, err = archive.blockio.Write(body)
	if err != nil {
		return errors.Wrap(err, "failed writing CBOR data")
	}

	return nil
}

func (archive *ArchiveWriter) AppendBytes(
	rtype format.RecordType,
	flags format.RecordFlags,
	meta any,
	data []byte) error {

	// Build preamble
	preamble := format.NewPreamble(rtype, flags, uint64(len(data)))

	hash, err := blake2b.New512(nil)
	if err != nil {
		// That's a problem
		return errors.Wrap(err, "Failed to initialize BLAKE2b hash")
	}

	// CBOR encode the metadata
	cborData, err := cbor.Marshal(meta)

	// pad the CBOR data out to 4K

	headerblob := make([]byte, format.BLOCK_SIZE)

	preambleBytes := preamble.ToBytes()

	copy(headerblob, preambleBytes)
	copy(headerblob[len(preambleBytes):], cborData)

	hash.Write(headerblob)
	hash.Write(data)
	checksum := hash.Sum(nil)

	if err != nil {
		return errors.Wrap(err, "Failed to marshal header data")
	}

	if copy(preamble.Checksum[:], checksum) != 64 {
		return errors.New("Couldn't copy checksum into preamble")
	}

	copy(headerblob, preambleBytes)

	if err = archive.blockio.Align(); err != nil {
		return errors.Wrap(err, "Failed to align to next block.")
	}

	if _, err = archive.blockio.Write(headerblob); err != nil {
		return errors.Wrap(err, "Failed to write header data.")
	}

	if _, err = archive.blockio.WriteWhole(data); err != nil {
		return errors.Wrap(err, "Failed to write body data")
	}

	return nil
}

func (archive *ArchiveWriter) AppendStream(path string, info fs.FileInfo, stream io.Reader) error {
	// write the file to the end of the archive.

	if info.Size() < int64(archive.MaxReadBuffer) {
		// We can read the whole thing into memory, get the whole file hash, etc.
	}

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

func (archive *ArchiveWriter) Close() error {
	return archive.blockio.Close()

}
