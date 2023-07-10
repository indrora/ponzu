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
	// Record type
	Rtype RecordType
	// Compression type of the body
	Compression CompressionType
	// Record flags
	Flags RecordFlags
	// Number of data-blocks that follow
	DataLen uint64
	// Number of bytes used in final data-block
	Modulo uint16
	// Checksum of data blocks
	DataChecksum [64]byte
	// Metadata Length
	MetadataLength uint16
	// checksum of the metadata
	MetadataChecksum [64]byte
}

func NewPreamble(
	rType RecordType,
	compression CompressionType,
	flags RecordFlags,
	length uint64,
	dataChecksum []byte,
	metadataLen uint16,
	metadataChecksum []byte) Preamble {

	bcount := uint64(0)
	modulo := uint16(0)

	if length == 0 {
		bcount = 0
		modulo = 0
	} else if length > 1 && length <= uint64(BLOCK_SIZE) {
		bcount = 1 // always a minimum of 1
		modulo = uint16(length)
	} else {
		bcount = 1 + (length / uint64(BLOCK_SIZE))
		modulo = uint16(length % uint64(BLOCK_SIZE))
	}

	return Preamble{
		Magic:        [6]byte{'P', 'O', 'N', 'Z', 'U', 0},
		Rtype:        rType,
		Compression:  compression,
		Flags:        flags,
		DataChecksum: [64]byte(dataChecksum),
		// computed fields
		DataLen:          bcount,
		Modulo:           modulo,
		MetadataLength:   metadataLen,
		MetadataChecksum: [64]byte(metadataChecksum),
	}
}

func (p *Preamble) ToBytes() []byte {

	b := new(bytes.Buffer)

	p.WritePreamble(b)

	return b.Bytes()

}

func (p *Preamble) WritePreamble(w io.Writer) error {

	if err := binary.Write(w, binary.BigEndian, *p); err != nil {
		return errors.Wrap(err, "failed to write preamble")
	}
	return nil
}

func ReadPreamble(r io.Reader) (*Preamble, error) {
	nPreamble := &Preamble{}
	err := binary.Read(r, binary.BigEndian, nPreamble)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read preamble")
	} else {
		return nPreamble, nil
	}

}

const BLOCK_SIZE uint64 = 4096

const (
	RECORD_TYPE_CONTROL     RecordType = 0
	RECORD_TYPE_FILE        RecordType = 1
	RECORD_TYPE_HARDLINK    RecordType = 2
	RECORD_TYPE_SYMLINK     RecordType = 3
	RECORD_TYPE_DIRECTORY   RecordType = 4
	RECORD_TYPE_ZDICTIONARY RecordType = 5
	RECORD_TYPE_OS_SPECIAL  RecordType = 126
	RECORD_TYPE_CONTINUE    RecordType = 127
)

const (
	RECORD_FLAG_NONE          RecordFlags = 0b00
	RECORD_FLAG_CONTROL_START RecordFlags = 0b1
	RECORD_FLAG_CONTROL_END   RecordFlags = 0b10
	RECORD_FLAG_CONTINUES     RecordFlags = 0b10
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
