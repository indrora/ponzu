package writer

import (
	"bufio"
	"bytes"
	"io"
	"io/fs"
	"os"

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

	return archive.AppendBytes(format.RECORD_TYPE_START, format.RECORD_FLAG_NONE, archiveHeader, nil)
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

func (archive *ArchiveWriter) AppendStream(stream io.Reader, recordType format.RecordType, flags format.RecordFlags, info any) error {
	readbuff := new(bytes.Buffer)

	reader := bufio.NewReaderSize(stream, int(archive.MaxReadBuffer))

	got, err := io.CopyN(readbuff, reader, int64(archive.MaxReadBuffer))

	if err != nil {
		return err
	}
	if got <= int64(archive.MaxReadBuffer) {
		return archive.AppendBytes(
			recordType, flags,
			info,
			readbuff.Bytes(),
		)
	} else {

		// See if we still have some amount of data left
		got, err = io.CopyN(readbuff, stream, int64(readGoal))
		if got < int64(readGoal) {
			return archive.AppendBytes(
				recordType, flags,
				info,
				readbuff.Bytes(),
			)
		} else {
			// Something else is left and we'll fill this function out later.
		}
	}

	return nil

}

func (archive *ArchiveWriter) AppendFileStream(stream io.Reader, fileInfo format.File) error {

	return nil
}

func (archive *ArchiveWriter) AppendFile(path string, source string, compression format.CompressionType) error {

	return nil
}

func (archive *ArchiveWriter) appendUncompressed(path string, source string) error {

	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	handle, err := os.Open(path)
	if err != nil {
		return err
	}
	defer handle.Close()
	err = archive.AppendStream(handle, format.RECORD_TYPE_FILE, format.RECORD_FLAG_NONE, format.File{
		Compressor: format.COMPRESSION_NONE,
		Name:       path,
		ModTime:    info.ModTime(),
	})
	return nil
}

func (archive *ArchiveWriter) appendCompressed(path string, source string, compression format.CompressionType) error {

	return nil
}

func (archive *ArchiveWriter) AppendDirectory(path string, info fs.FileInfo) error {

	err := archive.AppendBytes(format.RECORD_TYPE_DIRECTORY, format.RECORD_FLAG_NONE, format.Directory{
		File: format.File{Name: path,
			ModTime:    info.ModTime(),
			Compressor: format.COMPRESSION_NONE,
			Metadata:   map[string]any{},
		},
	}, nil)

	return err
}

func (archive *ArchiveWriter) AppendSymlink(path string, destination string, info fs.FileInfo) error {
	err := archive.AppendBytes(format.RECORD_TYPE_DIRECTORY, format.RECORD_FLAG_NONE, format.Symlink{
		Link: format.Link{
			File: format.File{Name: path,
				ModTime:    info.ModTime(),
				Compressor: format.COMPRESSION_NONE,
				Metadata:   map[string]any{},
			},
			Target: destination,
		},
	}, nil)

	return err

}

func (archive *ArchiveWriter) Close() error {
	return archive.blockio.Close()
}
