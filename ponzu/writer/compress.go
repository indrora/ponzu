package writer

import (
	"bytes"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
)

func (archive *ArchiveWriter) getCompressedChunk(data []byte, compressor format.CompressionType) ([]byte, error) {

	buf := new(bytes.Buffer)

	writer, err := archive.GetCompressor(buf, compressor)

	if err != nil {
		return nil, err
	}
	_, err = writer.Write(data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

func (archive *ArchiveWriter) GetCompressor(writer io.Writer, compressor format.CompressionType) (io.Writer, error) {

	switch compressor {
	case format.COMPRESSION_NONE:
		return writer, nil
	case format.COMPRESSION_BROTLI:
		return brotli.NewWriter(writer), nil
	case format.COMPRESSION_ZSTD:
		return zstd.NewWriter(writer)
	default:
		return nil, errors.New("Unknown compressor")
	}

}
