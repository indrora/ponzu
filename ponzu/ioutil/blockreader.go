package ioutil

import (
	"bytes"
	"io"
)

type BlockReader struct {
	reader    io.Reader
	ChunkSize int64
	buffer    *bytes.Buffer
}

func NewBlockReader(reader io.Reader, chunkSize int64) *BlockReader {
	return &BlockReader{
		reader:    reader,
		ChunkSize: chunkSize,
		//buffer:    new(bytes.Buffer),
	}
}

func (br *BlockReader) ReadBlock() ([]byte, error) {

	buffer := new(bytes.Buffer)
	n, err := io.CopyN(buffer, br.reader, int64(br.ChunkSize))
	if err == io.EOF {
		if n == 0 {
			return nil, io.EOF
		}
		return buffer.Bytes(), io.EOF
	}
	if n < br.ChunkSize {
		return buffer.Bytes(), io.EOF
	}
	return buffer.Bytes(), nil
}
