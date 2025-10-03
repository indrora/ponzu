package ioutil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type BlockReader struct {
	reader       *bufio.Reader
	ChunkSize    uint64
	realignBytes uint64
}

func NewBlockReader(reader io.Reader, chunkSize uint64) *BlockReader {
	return &BlockReader{
		reader:    bufio.NewReaderSize(reader, int(chunkSize)),
		ChunkSize: chunkSize,
	}
}

// This is for convenience
func (br *BlockReader) Read(b []byte) (int, error) {

	read, err := br.reader.Read(b)

	br.realignBytes += uint64(read)

	return read, err
}

func (br *BlockReader) Realign() error {

	if br.realignBytes == 0 {
		return nil
	} else if br.realignBytes > br.ChunkSize {
		br.realignBytes = br.realignBytes % br.ChunkSize
	}

	if br.realignBytes > 0 {
		// Read out the remaining bytes

		discardable := br.ChunkSize - br.realignBytes
		fmt.Printf("Discarding %d bytes \n", discardable)
		_, err := io.CopyN(io.Discard, br.reader, int64(discardable))
		if err != nil && err != io.EOF {
			return err
		}
		br.realignBytes = 0
		return err
	}

	return nil

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
	if uint64(n) < br.ChunkSize {
		return buffer.Bytes(), io.EOF
	} else if _, err = br.reader.Peek(1); err == io.EOF {
		return buffer.Bytes(), io.EOF
	}

	return buffer.Bytes(), err

}
