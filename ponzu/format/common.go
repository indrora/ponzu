package format

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

type PTimestamp uint64

/*

This preamble is in the first little bit of the record.

*/

const (
	PREAMBLE_STRING = "PONZU"
	PONZU_VERSION   = 1
)

var (
	PREAMBLE_BYTES = []byte{'P', 'O', 'N', 'Z', 'U', 0}
)

type RecordType uint8
type RecordFlags uint16
type Preamble struct {
	// Magic value, must be PREAMBLE_STRING
	Magic [6]byte
	// Record type (0 = SOA, etc. )
	Rtype RecordType
	// Record flags (Half, Streamed, etc.)
	Flags RecordFlags
	// Number of data-blocks that follow
	DataLen uint64
	// Number of bytes used in final data-block
	Modulo uint16
	// Checksum of data blocks
	Checksum [64]byte
}

func NewPreamble(rType RecordType, flags RecordFlags, length uint64) Preamble {

	bcount := uint64(0)
	modulo := uint16(0)

	if length == 0 {
		bcount = 0
		modulo = 0
	} else if flags&RECORD_FLAG_HALF == RECORD_FLAG_HALF {
		// check that length < 1/2 BLOCK_SIZE
		if length > uint64(BLOCK_SIZE)/2 {
			flags ^= RECORD_FLAG_HALF
		}
		bcount = 0
		modulo = uint16(length)

	} else if length > 1 && length < uint64(BLOCK_SIZE) {
		bcount = 1 // always a minimum of 1
		modulo = uint16(length)
	} else {
		bcount = 1 + (length / uint64(BLOCK_SIZE))
		modulo = uint16(length % uint64(BLOCK_SIZE))
	}

	return Preamble{
		Magic:    [6]byte{'P', 'O', 'N', 'Z', 'U', 0},
		Rtype:    rType,
		Flags:    flags,
		Checksum: [64]byte{0},
		// computed fields
		DataLen: bcount,
		Modulo:  modulo,
	}
}

func (p *Preamble) ToBytes() []byte {

	b := new(bytes.Buffer)

	p.WritePreamble(b)

	return b.Bytes()

}

func (p *Preamble) WritePreamble(w io.Writer) error {

	if err := binary.Write(w, binary.BigEndian, p.Magic); err != nil {
		return errors.Wrap(err, "failed to write preamble")
	}
	if err := binary.Write(w, binary.BigEndian, p.Rtype); err != nil {
		return errors.Wrap(err, "failed to write preamble")
	}
	if err := binary.Write(w, binary.BigEndian, p.Flags); err != nil {
		return errors.Wrap(err, "failed to write preamble")
	}
	if err := binary.Write(w, binary.BigEndian, p.DataLen); err != nil {
		return errors.Wrap(err, "failed to write preamble")
	}
	if err := binary.Write(w, binary.BigEndian, p.Modulo); err != nil {
		return errors.Wrap(err, "failed to write preamble")
	}
	if err := binary.Write(w, binary.BigEndian, p.Checksum); err != nil {
		return errors.Wrap(err, "failed to write preamble")
	}
	return nil
}

const BLOCK_SIZE int64 = 4096

const (
	RECORD_TYPE_START       RecordType = 0
	RECORD_TYPE_FILE        RecordType = 1
	RECORD_TYPE_HARDLINK    RecordType = 2
	RECORD_TYPE_SYMLINK     RecordType = 3
	RECORD_TYPE_DIRECTORY   RecordType = 4
	RECORD_TYPE_ZDICTIONARY RecordType = 5
	RECORD_TYPE_OS_SPECIAL  RecordType = 126
	RECORD_TYPE_CONTINUE    RecordType = 127
)

const (
	RECORD_FLAG_NONE      RecordFlags = 0b00
	RECORD_FLAG_HALF      RecordFlags = 0b01
	RECORD_FLAG_STREAMED  RecordFlags = 0b10
	RECORD_FLAG_CONTINUES RecordFlags = 0b10
	RECORD_FLAG_STAMPED   RecordFlags = 0b100
)

type CompressionType uint8

const (
	COMPRESSION_NONE   CompressionType = 0
	COMPRESSION_ZSTD   CompressionType = 1
	COMPRESSION_BROTLI CompressionType = 3
)

const (
	HOST_OS_GENERIC string = "universe"
	HOST_OS_LINUX   string = "linux"
	HOST_OS_UNIX    string = "unix"
	HOST_OS_SELINUX string = "selinux"
	HOST_OS_NT      string = "winnt"
	HOST_OS_DARWIN  string = "darwin"
	HOST_OS_POSIX   string = "posix"
)
