package ioutil

import (
	"hash"
	"io"
)

type HashWriter struct {
	writer io.Writer
	hasher hash.Hash
}

func NewHashWriter(dest io.Writer, hasher hash.Hash) *HashWriter {
	return &HashWriter{
		writer: dest,
		hasher: hasher,
	}
}

func (w *HashWriter) Write(b []byte) (int, error) {
	w.hasher.Write(b)
	k, err := w.writer.Write(b)
	if err != nil {
		return 0, err
	}
	return k, nil
}

func (w *HashWriter) Sum() []byte {
	return w.hasher.Sum(nil)
}
