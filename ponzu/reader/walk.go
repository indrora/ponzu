package reader

import (
	"io"

	"github.com/indrora/ponzu/ponzu/format"
)

type WalkFunc func(*format.Preamble, *format.RecordInfo) error

func (archive *Reader) Walk(fn WalkFunc) error {

	for {
		preamble, info, err := archive.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := fn(preamble, info); err != nil {
			return err
		}
	}

	return nil
}
