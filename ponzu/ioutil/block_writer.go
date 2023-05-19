package ioutil

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
)

type BlockWriter struct {
	writer              io.Writer
	writtenSinceRealign uint64
	bsize               uint64
}

func NewBlockWriter(destination io.Writer, blockSize uint64) *BlockWriter {
	return &BlockWriter{
		writer:              destination,
		bsize:               blockSize,
		writtenSinceRealign: 0,
	}
}

// Writes bytes to the
func (k *BlockWriter) Write(p []byte) (n int, err error) {
	written, err := k.writer.Write(p)
	k.writtenSinceRealign += uint64(written)
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

	if k.writtenSinceRealign > uint64(k.bsize) {
		for k.writtenSinceRealign > uint64(k.bsize) {
			k.writtenSinceRealign -= uint64(k.bsize)
		}
	}
	if k.writtenSinceRealign != 0 {
		// Write out the remaining portion of a block
		toWrite := k.bsize - k.writtenSinceRealign

		empty := bytes.Repeat([]byte{0}, int(toWrite))

		_, err := k.writer.Write(empty)
		if err != nil {
			return errors.Wrap(err, "Failed to finish out block")
		}
	}
	k.writtenSinceRealign = 0
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
