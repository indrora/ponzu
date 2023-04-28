package reader

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/fxamacker/cbor/v2"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/ioutil"
)

// The reader is much simpler than the writer.

type ReaderState int



var (
	ErrExpectedHeader    = errors.New("expected a header, got something else")
	ErrUnexpectedControl = errors.New("unexpected control message")
	ErrUnknownRecordType = errors.New("unknown record type")
	ErrMalformedHash     = errors.New("hash does not match")
	ErrState        = errors.New("tried reading body before you got a header")
)


const (
	STATE_EMPTY ReaderState = 0
	STATE_OK    ReaderState = 1
	STATE_BODY  ReaderState = 2
	
)


type Reader struct {
	stream       *ioutil.BlockReader
	lastPreamble *format.Preamble
	state			ReaderState 
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{
		stream:       ioutil.NewBlockReader(reader, format.BLOCK_SIZE),
		lastPreamble: nil,
	}
}

func (reader *Reader) Next() (*format.Preamble, interface{}, error) {

	if reader.lastPreamble != nil {
		// we have a previous header!

		for idx := uint64(0); idx < reader.lastPreamble.DataLen; idx++ {
			_, err := reader.stream.ReadBlock()
			if err != nil {
				return nil, nil, err
			}
		}
	}

	block, err := reader.stream.ReadBlock()

	if err != nil {
		return nil, nil, err
	}
	hReader := bytes.NewReader(block)
	mPreamble := &format.Preamble{}

	if err = binary.Read(hReader, binary.BigEndian, mPreamble); err != nil {
		return nil, nil, errors.Join(err, ExpectedHeader)
	}


	// do some quick checks on the preamble

	switch(mPreamble.Rtype) {
		
	case format.RECORD_TYPE_CONTROL:
	case format.RECORD_TYPE_DIRECTORY:
	case format.RECORD_TYPE_HARDLINK:
	case format.RECORD_TYPE_SYMLINK:
		if mPreamble.DataLen != 0 {
			return errors.Join(ExpectedHeader, errors.New("Invalid block "))
		}

	default:
		// noop
	}

	cborDecoder := cbor.NewDecoder(hReader)

	var metadata interface{}

	if err = cborDecoder.Decode(&metadata); err != nil {
		return nil, nil, errors.join(err, ExpectedHeader)
	}

	reader.lastPreamble = mPreamble

	return mPreamble, metadata, nil

}

func (reader *Reader) HasBody() bool {
	if reader.lastPreamble != nil {
		return reader.lastPreamble.DataLen != 0
	} else {
		return false
	}
}

func (reader *Reader) GetBody() ([]byte, error) {

	if 

	// noop for now
	return nil, nil
}

func (reader *Reader) CopyTo(writer io.Writer, validate bool) error {
	// noop
	return nil
}

func (reader *Reader) CopyAll(writer io.Writer, validate bool) error {
	// noop

	return nil
}
