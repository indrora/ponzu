package writer

import (
	"bytes"

	"github.com/andybalholm/brotli"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/klauspost/compress/zstd"
	"github.com/pkg/errors"
)

func (archive *ArchiveWriter) getCompressedChunk(data []byte, compressor format.CompressionType) ([]byte, error) {

	switch compressor {
	case format.COMPRESSION_NONE:
		return data, nil
	case format.COMPRESSION_BROTLI:
		buf := new(bytes.Buffer)
		comp := brotli.NewWriter(buf)
		_, err := comp.Write(data)
		if err != nil {
			return nil, err
		}
		comp.Close()
		return buf.Bytes(), nil
	case format.COMPRESSION_ZSTD:
		buf := new(bytes.Buffer)

		cctx, err := zstd.NewWriter(buf)
		if err != nil {
			return nil, err
		}
		cctx.Write(data)

		cctx.Close()

		return buf.Bytes(), nil
	default:
		return nil, errors.New("unkonwn compressor")
	}
}
