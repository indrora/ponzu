package ioutil

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

type CompressWriter interface {
	// Compress reads from a Reader until there is no more to read,
	// compresses it into the writer,
	// and returns the bytes read, bytes written, and any error.
	Copy(io.Writer, io.Reader) (int64, error)
}

type CopyWriter struct {
	CompressWriter
}

func (compressor CopyWriter) Copy(reader io.Reader, writer io.Writer) (int64, error) {
	return io.Copy(writer, reader)

}

type ZstdWriter struct {
	Dictionary []byte
}

func (compressor ZstdWriter) Copy(reader io.Reader, writer io.Writer) (int64, error) {

	zWriter := (*zstd.Encoder)(nil)
	err := (error)(nil)
	if compressor.Dictionary == nil {

		zWriter, err = zstd.NewWriter(writer)
		if err != nil {
			return 0, err
		}
	} else {
		zWriter, err = zstd.NewWriter(writer, zstd.WithEncoderDict(compressor.Dictionary))
		if err != nil {
			return 0, err
		}
	}

	return zWriter.ReadFrom(reader)
}

type BrotliWriter struct{}

func (compressor BrotliWriter) Copy(reader io.Reader, writer io.Writer) (int64, error) {
	return 0, nil
}
