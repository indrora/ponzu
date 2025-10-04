package reader

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/ioutil"
	"golang.org/x/crypto/blake2b"
)

// The reader is much simpler than the writer.

var (
	ErrExpectedHeader    = errors.New("expected a header, got something else")
	ErrUnexpectedData    = errors.New("unexpected data length for record type")
	ErrExpectedContinue  = errors.New("expected continue, got other")
	ErrUnexpectedControl = errors.New("unexpected control message")
	ErrUnknownRecordType = errors.New("unknown record type")
	ErrHashMismatch      = errors.New("hash does not match")
	ErrState             = errors.New("tried reading body before you got a header")
	ErrWalk              = errors.New("walk function returned non-nil error")
)

// A Reader is the mechanism used to read the archive format.
//
// QUIRKS:
//
//   - the Next() call will silently eat zstandard dictionaries
//   - There's some calls to fmt.Println here that need cleaned up.
type Reader struct {
	stream            *ioutil.BlockReader // Underlying block reader
	currentPreamble   *format.Preamble    // Most recently read preamble
	currentRecordinfo *format.RecordInfo  // Current Record information.
	bodyRecordBytes   int64               // Body bytes left to read (not including modulus)
	zstdDict          []byte              // Current zstd dictionary
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{
		stream:            ioutil.NewBlockReader(reader, format.BLOCK_SIZE),
		currentPreamble:   nil,
		currentRecordinfo: nil,
		zstdDict:          nil,
		bodyRecordBytes:   0,
	}
}

func (reader *Reader) Next() (rPreamble *format.Preamble, rInfo *format.RecordInfo, err error) {

	if reader.currentPreamble != nil {
		// Toss away the rest of the content
		reader.CopyTo(io.Discard, false)
		reader.currentPreamble = nil
	}

	// make sure we're on a block boundary
	reader.stream.Realign()

	rPreamble = &format.Preamble{}

	preamblebytes := make([]byte, binary.Size(*rPreamble))
	i, e := reader.stream.Read(preamblebytes)

	if e != nil {
		if e == io.EOF {
			return nil, nil, io.EOF
		}
		return nil, nil, ErrExpectedHeader
	}

	if i != binary.Size(*rPreamble) {
		return nil, nil, ErrExpectedHeader
	}

	breader := bytes.NewReader(preamblebytes)

	if err = binary.Read(breader, binary.BigEndian, rPreamble); err != nil {
		return nil, nil, errors.Join(err, ErrExpectedHeader)
	}

	if !bytes.Equal(rPreamble.Magic[:], format.PREAMBLE_BYTES[:]) {
		return nil, nil, ErrExpectedHeader
	}

	// Copy out the record information bock.
	cborData := new(bytes.Buffer)
	n, err := io.CopyN(cborData, reader.stream, int64(rPreamble.InfoLength))

	if err == io.EOF {
		return nil, nil, io.EOF
	}

	// Check that we read the right amount of information.
	if n != int64(rPreamble.InfoLength) {
		return rPreamble, nil, fmt.Errorf("%w: Tried reading %v preamble bytes, got %v!", err, rPreamble.InfoLength, n)
	} else if err != nil {
		return nil, nil, err
	}

	cborDataBytes := cborData.Bytes()
	metaHashCheck := blake2b.Sum512(cborDataBytes)

	if !bytes.Equal(metaHashCheck[:], rPreamble.InfoChecksum[:]) {
		return nil, nil, fmt.Errorf("%w: record information checksum failed, expected %x, got %x ", ErrHashMismatch, rPreamble.InfoChecksum, metaHashCheck)
	}

	if len(cborDataBytes) > 0 {
		rInfo, err = UnmarshalRecordInfo(rPreamble, cborDataBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal record information: %w", err)
		}
	}

	// Realign the reader to the start of the data (or next record)
	reader.stream.Realign()

	reader.bodyRecordBytes = int64(rPreamble.DataLen * format.BLOCK_SIZE)

	// There are a few edge cases that need to be handled at this time.
	switch rPreamble.Rtype {
	case format.RECORD_TYPE_CONTROL:
	case format.RECORD_TYPE_DIRECTORY:
	case format.RECORD_TYPE_HARDLINK:
	case format.RECORD_TYPE_SYMLINK:
	case format.RECORD_TYPE_OS_SPECIAL:
		if rPreamble.DataLen != 0 {
			return rPreamble, rInfo, fmt.Errorf("%w: expected 0, got %v", ErrUnexpectedData, rPreamble.DataLen)
		}
	case format.RECORD_TYPE_ZDICTIONARY:
		// Special case: we are going to consume the zstd dictionary and then return the next frame afterwards
		// TODO: allow for this to just be passed along.
		buff := new(bytes.Buffer)
		err := reader.CopyAll(buff, true)
		if err != nil && err != io.EOF {
			return nil, nil, err
		} else {
			reader.zstdDict = buff.Bytes()
			return reader.Next()
		}

	default:
		// No special handling.
	}

	reader.currentPreamble = rPreamble
	reader.currentRecordinfo = rInfo

	return rPreamble, rInfo, nil

}

func (reader *Reader) HasBody() bool {

	if reader.currentPreamble != nil {
		return reader.currentPreamble.DataLen != 0
	} else {
		return false
	}
}

func (reader *Reader) Validate() (bool, error) {
	if !reader.HasBody() {
		return true, nil
	}

	// Hack: since the checksum is calculated after compression, we don't care.
	reader.currentPreamble.Compression = format.COMPRESSION_NONE

	err := reader.CopyAll(io.Discard, true)
	if err == ErrHashMismatch {
		return false, err
	}
	// CopyTo returns io.EOF when finished.
	if err != io.EOF {
		return false, err
	}
	return true, nil
}

func (reader *Reader) GetBody(validate bool) ([]byte, error) {

	// if there is no body, we clean up the header and leave.

	// decompress it into the appropriate buffer.
	body := new(bytes.Buffer)

	err := reader.CopyTo(body, validate)

	if err != nil && err != io.EOF {
		return nil, err
	}

	err2 := reader.stream.Realign()
	return body.Bytes(), errors.Join(err, err2)
}

func (reader *Reader) CopyTo(writer io.Writer, validate bool) error {

	// if there is no body, we clean up the header and leave.

	var err error

	if !reader.HasBody() {
		reader.stream.Realign()
		reader.currentPreamble = nil
		return nil
	}

	// Otherwise, we're going to fill up our buffer.
	bodyLen := (reader.currentPreamble.DataLen * format.BLOCK_SIZE)

	if reader.currentPreamble.Modulo != 0 {
		bodyLen = bodyLen - (format.BLOCK_SIZE - uint64(reader.currentPreamble.Modulo))
	}

	// Get a limited reader
	dataReader := io.LimitReader(reader.stream, int64(bodyLen))

	// set up the tee: This allows us to compute the checksum in-situ, while the read is happening
	// at no performance penalty.
	hash, _ := blake2b.New512(nil)
	// tee from the limited reader to the hash function glub glub
	tee := io.TeeReader(dataReader, hash)
	// Wrap it in our decompression function (in the simple case, this is null, otherwise this is a zstd/brotli decompressor)
	dataReader, err = reader.getDecompressor(tee, reader.currentPreamble.Compression)

	if err != nil {
		return err
	}

	_, err = io.Copy(writer, dataReader)

	if err != nil && err != io.EOF {
		// something terrible has happened.
		return err

	}

	checksum := hash.Sum(nil)

	alignerr := reader.stream.Realign()
	// if we've been asked to validate the checksum, do it now

	if validate {
		if !bytes.Equal(checksum, reader.currentPreamble.DataChecksum[:]) {
			reader.currentPreamble = nil
			return ErrHashMismatch
		}
	}

	reader.currentPreamble = nil
	return errors.Join(err, alignerr)
}

func (reader *Reader) CopyAll(writer io.Writer, validate bool) error {
more:

	if reader.bodyRecordBytes == 0 {
		return io.EOF
	}

	continues := reader.currentPreamble.Flags&format.RECORD_FLAG_CONTINUES == format.RECORD_FLAG_CONTINUES

	err := reader.CopyTo(writer, validate)
	if err != nil && err != io.EOF {
		return err
	}

	if continues {
		tPre, _, err := reader.Next()
		if err != nil {
			return err
		} else if tPre.Rtype != format.RECORD_TYPE_CONTINUE || tPre.Flags&format.RECORD_FLAG_CONTINUES == 0 {
			return ErrExpectedContinue
		}
		goto more
	}

	return nil
}
