package writer

import (
	"bufio"
	"bytes"
	"io"

	"github.com/indrora/ponzu/ponzu/format"
)

/// Stream stuff

func (archive *ArchiveWriter) AppendStream(stream io.Reader, recordType format.RecordType, flags format.RecordFlags, info any) error {
	readbuff := new(bytes.Buffer)

	reader := bufio.NewReaderSize(stream, int(archive.MaxReadBuffer))

	got, err := io.CopyN(readbuff, reader, int64(archive.MaxReadBuffer))

	if err != nil {
		return err
	}
	if got <= int64(archive.MaxReadBuffer) {
		return archive.AppendBytes(
			recordType, flags,
			info,
			readbuff.Bytes(),
		)
	} else {

		// See if we still have some amount of data left
		got, err = io.CopyN(readbuff, stream, int64(readGoal))
		if got < int64(readGoal) {
			return archive.AppendBytes(
				recordType, flags,
				info,
				readbuff.Bytes(),
			)
		} else {
			// Something else is left and we'll fill this function out later.
		}
	}

	return nil

}

func (archive *ArchiveWriter) AppendFileStream(stream io.Reader, fileInfo format.File) error {

	return nil
}
