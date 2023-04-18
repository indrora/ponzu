package ioutil

import (
	"bufio"
	"bytes"
	"io"
)

type BlockReader struct {
	reader    *bufio.Reader
	ChunkSize int64
}

func NewBlockReader(reader io.Reader, chunkSize int64) *BlockReader {
	return &BlockReader{
		reader:    bufio.NewReaderSize(reader, int(chunkSize)),
		ChunkSize: chunkSize,
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
	} else if _, err = br.reader.Peek(1); err == io.EOF {
		return buffer.Bytes(), io.EOF
	}

	return buffer.Bytes(), err

}
