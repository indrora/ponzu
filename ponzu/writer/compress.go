package writer

import (
	"bytes"

	"github.com/google/brotli/go/cbrotli"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
)

func (archive *ArchiveWriter) getCompressedChunk(data []byte, compressor format.CompressionType) ([]byte, error) {
	switch compressor {
	case format.COMPRESSION_NONE:
		return data, nil
	case format.COMPRESSION_BROTLI:
		return cbrotli.Encode(data, cbrotli.WriterOptions{})
	case format.COMPRESSION_ZSTD:
		// slightly harder to handle. There's one little hiccup:
		// if the archive has a current zstd dictionary, we should use that
		// but for now, we're not going to care.

		buf := new(bytes.Buffer)
		writer, err := zstd.NewWriter(buf)
		if err != nil {
			return nil, err
		}
		_, err = writer.Write(data)
		if err != nil {
			return nil, err
		}

		writer.Flush()
		writer.Close()
		return buf.Bytes(), nil
	default:
		return nil, errors.New("Unknown compression format.")
	}
}
