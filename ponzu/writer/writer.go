package writer

import (
	"bytes"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/ioutil"
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

func (archive *ArchiveWriter) AppendStart(prefix string, comment string) error {
	// write the initial header to the file.

	// This is the CBOR portion.
	archiveHeader := format.StartOfArchive{
		Version: format.PONZU_VERSION,
		Host:    format.HOST_OS_GENERIC,
		Prefix:  prefix,
		Comment: comment,
	}

	return archive.AppendBytes(format.RECORD_TYPE_CONTROL, format.RECORD_FLAG_CONTROL_START, format.COMPRESSION_NONE, archiveHeader, nil)
}

func (archive *ArchiveWriter) AppendEnd() error {
	return archive.AppendBytes(format.RECORD_TYPE_CONTROL, format.RECORD_FLAG_CONTROL_END, format.COMPRESSION_NONE, nil, nil)
}

// AppendBytes adds a raw, uncompressed block of data to the end of the archive.
// This includes the header and relevant body (`recordInfo`)
func (archive *ArchiveWriter) AppendBytes(
	rtype format.RecordType,
	flags format.RecordFlags,
	compression format.CompressionType,
	recordInfo any,
	data []byte) error {

	// Build preamble

	// we can safely assume that if there is data, we can get the length of the data.
	dlen := uint64(0)
	if data != nil {
		dlen = uint64(len(data))
	}

	// we may or may not have CBOR data, depending on if we have any metadata to append.

	var err error
	var cborData []byte

	if recordInfo != nil {
		// CBOR encode the metadata
		cborData, err = cbor.Marshal(recordInfo)

		if err != nil {
			return errors.Wrap(err, "Failed to marshal metadata to CBOR.")
		}
	} else {
		cborData = []byte{}
	}

	metadataChecksum := blake2b.Sum512(cborData)
	metadataLengh := len(cborData)

	if data != nil {

		data, err = archive.getCompressedChunk(data, compression)

		if err != nil {
			return errors.Wrap(err, "failed to compress data")
		}
	} else {
		data = []byte{}
	}

	bodyChecksum := blake2b.Sum512(data)

	headerbuf := new(bytes.Buffer)

	preamble := format.NewPreamble(rtype, compression, flags, dlen, bodyChecksum[:], uint16(metadataLengh), metadataChecksum[:])

	// Write the preamble out
	preamble.WritePreamble(headerbuf)
	// Now write the cbor data to the buffer
	headerbuf.Write(cborData)

	// Write the full record header to the block io -- this pads out to the next 4K block.
	archive.blockio.WriteWhole(headerbuf.Bytes())
	// if we have data, append it here.
	if data != nil {
		archive.blockio.WriteWhole(data)
	}

	return nil
}

func (archive *ArchiveWriter) AppendStream(rtype format.RecordType, flags format.RecordFlags, compression format.CompressionType, recordInfo any, stream io.Reader) error {

	chunkReader := ioutil.NewBlockReader(stream, archive.MaxReadBuffer/2)

	// Read at least the first chunk

	chunk, err := chunkReader.ReadBlock()

	if err == io.EOF {
		// There is only one block, we're cool
		return archive.AppendBytes(rtype, flags, compression, recordInfo, chunk)
	} else if err != nil {
		return errors.Wrap(err, "failed to read block from underlying stream")
	} else {
		// Tick on the CONTINUES flag

		flags |= format.RECORD_FLAG_CONTINUES
		if err = archive.AppendBytes(rtype, flags, compression, recordInfo, chunk); err != nil {
			return errors.Wrap(err, "Failed to write first chunk in continue chain")
		} else {
			for {

				chunk, err = chunkReader.ReadBlock()
				if err != nil && err != io.EOF {
					return errors.Wrap(err, "failed to read block from underlying stream")
				} else if err == io.EOF {
					return archive.AppendBytes(format.RECORD_TYPE_CONTINUE, format.RECORD_FLAG_NONE, compression, nil, chunk)
				} else {
					err = archive.AppendBytes(format.RECORD_TYPE_CONTINUE, format.RECORD_FLAG_CONTINUES, compression, nil, chunk)
					if err != nil {
						return errors.Wrap(err, "failed to write continuation block.")
					}
				}

			}
		}

	}

}

func (archive *ArchiveWriter) Close() error {
	return archive.blockio.Close()
}
