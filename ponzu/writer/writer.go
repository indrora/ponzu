package writer

import (
	"bytes"
	"io"
	"os"

	"github.com/fxamacker/cbor/v2"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/format/metadata"
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

func (archive *ArchiveWriter) AppendSOA(prefix string, comment string) error {
	// write the initial header to the file.

	// This is the CBOR portion.
	archiveHeader := format.StartOfArchive{
		Version: format.PONZU_VERSION,
		Host:    format.HOST_OS_GENERIC,
		Prefix:  prefix,
		Comment: comment,
	}

	return archive.AppendBytes(format.RECORD_TYPE_CONTROL, format.RECORD_FLAG_CONTROL_START, archiveHeader, nil)
}

func (archive *ArchiveWriter) AppendBytes(
	rtype format.RecordType,
	flags format.RecordFlags,
	recordInfo any,
	data []byte) error {

	// Build preamble

	dlen := uint64(0)
	if data != nil {
		dlen = uint64(len(data))
	}

	preamble := format.NewPreamble(rtype, flags, dlen)

	hash, err := blake2b.New512(nil)
	if err != nil {
		// That's a problem
		return errors.Wrap(err, "Failed to initialize BLAKE2b hash")
	}

	// CBOR encode the metadata
	cborData, err := cbor.Marshal(recordInfo)

	if err != nil {
		return errors.Wrap(err, "Failed to marshal metadata to CBOR.")
	}

	// pad the CBOR data out to 4K

	headerbuf := new(bytes.Buffer)

	preamble.WritePreamble(headerbuf)
	headerbuf.Write(cborData)
	// now, pad it out

	// This is dumb, but it works
	for {
		if headerbuf.Len() < int(format.BLOCK_SIZE) {
			headerbuf.WriteByte(0)
		} else {
			break
		}
	}

	// Hash the thing
	recordbytes := headerbuf.Bytes()
	_, err = hash.Write(recordbytes)
	if err != nil {
		return errors.Wrap(err, "Failed to hash header content")
	}
	if data != nil {
		_, err = hash.Write(data)
		if err != nil {
			return errors.Wrap(err, "failed to hash body content")
		}
	}

	copy(preamble.Checksum[0:], hash.Sum(nil))
	copy(recordbytes[0:], preamble.ToBytes())

	if _, err = archive.blockio.Write(recordbytes); err != nil {
		return errors.Wrap(err, "failed to write to underlying stream")
	}

	if data != nil {

		if _, err = archive.blockio.WriteWhole(data); err != nil {
			return errors.Wrap(err, "failed to write to underlying stream")
		}
	}
	return nil
}

func (archive *ArchiveWriter) AppendFile(path string, source string, compression format.CompressionType) error {

	// Get the stat of it
	stat, err := os.Stat(source)

	if err != nil {
		return err
	}

	// Construct a file record for it

	szui := (uint64)(stat.Size())

	info := format.File{
		Name:       path,
		ModTime:    stat.ModTime(),
		Compressor: compression,
		Metadata: metadata.CommonMetadata{
			FileSize: &szui,
		},
	}

	if stat.Size() <= int64(archive.MaxReadBuffer) {
		// Hot dang we can just read the whole thing in
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		data, err = archive.getCompressedChunk(data, compression)
		if err != nil {
			return err
		}
		return archive.AppendBytes(format.RECORD_TYPE_FILE, format.RECORD_FLAG_NONE, info, data)
	} else {

		// We're going to read chunks

		handle, err := os.OpenFile(source, os.O_RDONLY, os.ModeExclusive)
		if err != nil {
			return err
		}
		defer handle.Close()

		chunkReader := ioutil.NewBlockReader(handle, int64(archive.MaxReadBuffer)/2)
		chunk, chunkErr := chunkReader.ReadBlock()

		if chunk == nil && chunkErr != io.EOF {
			// Somehow, we hit the end of the file
		} else if chunkErr == io.EOF {
			// We've hit the end of the file.

		} else {
			// dump dump dump
		}
		_, err = compression_func.Write(chunk)

	}

	return nil
}

func (archive *ArchiveWriter) Close() error {
	err := archive.AppendBytes(format.RECORD_TYPE_CONTROL, format.RECORD_FLAG_CONTROL_END, nil, nil)
	if err != nil {
		return err
	}
	return archive.blockio.Close()
}
