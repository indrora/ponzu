package ioutil

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
)

type BlockWriter struct {
	writer io.Writer
	modulo int64
	bsize  int64
}

func NewBlockWriter(destination io.Writer, blockSize int64) *BlockWriter {
	return &BlockWriter{
		writer: destination,
		bsize:  blockSize,
		modulo: 0,
	}
}

func (k *BlockWriter) Write(p []byte) (n int, err error) {
	written, err := k.writer.Write(p)
	k.modulo = int64(written) % k.bsize
	return written, err
}

func (k *BlockWriter) WriteWhole(p []byte) (n int, err error) {
	n, err = k.Write(p)
	if err != nil {
		return n, errors.Wrap(err, "Failed to write block")
	}
	err = k.Align()
	return n, err
}

func (k *BlockWriter) Align() error {
	if k.modulo != 0 {
		// Write out the remaining portion of a block
		_, err := k.writer.Write(bytes.Repeat([]byte{0}, (int)(k.bsize-k.modulo)))
		if err != nil {
			return errors.Wrap(err, "Failed to finish out block")
		}
	}
	k.modulo = 0
	return nil
}

func (k *BlockWriter) Close() error {
	k.Align()

	if closer, ok := k.writer.(io.Closer); ok {
		return closer.Close()
	} else {
		return nil
	}
}
