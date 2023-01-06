package writer

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/indrora/ponzu/ponzu/format"
)

/// Stream stuff

func (archive *ArchiveWriter) AppendStream(stream io.Reader, recordType format.RecordType, flags format.RecordFlags, info any) error {
	readbuff := new(bytes.Buffer)

	reader := bufio.NewReaderSize(stream, int(archive.MaxReadBuffer))

	readErr := errors.Unwrap(nil)

	for readErr == nil {
		got, readErr := io.CopyN(readbuff, reader, int64(archive.MaxReadBuffer))

		if readErr == io.EOF {
			// There are no more bytes to read.
		}

		if got == 0 {
			// We read zero bytes.
		}

	}

	return nil

}

func (archive *ArchiveWriter) AppendFileStream(stream io.Reader, fileInfo format.File) error {

	return nil
}
