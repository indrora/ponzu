package reader

import (
	"io"

	"github.com/indrora/ponzu/ponzu/format"
)

type WalkFunc func(format.Preamble, format.RecordBase) error

func (archive *Reader) Walk(fn WalkFunc) error {

	for {
		preamble, data, err := archive.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		record := UnmarshalMetadata(preamble, data)
		if err != nil {
			return err
		}

		if err := fn(*preamble, record); err != nil {
			return err
		}
	}

	return nil
}
