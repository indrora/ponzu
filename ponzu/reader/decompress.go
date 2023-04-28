package reader

import (
	"errors"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/klauspost/compress/zstd"
)

var (
	UnknownCompressionError = errors.New("unknown compression")
)

func (reader *Reader) getDecompressor(compressedReader io.Reader, dcType format.CompressionType) (io.Reader, error) {

	switch dcType {
	case format.COMPRESSION_NONE:
		return compressedReader, nil // no compression = passthru
	case format.COMPRESSION_BROTLI:
		return brotli.NewReader(compressedReader), nil
	case format.COMPRESSION_ZSTD:
		return zstd.NewReader(compressedReader)
	default:
		return nil, UnknownCompressionError
	}

}
