package format

type PTimestamp uint64

/*

The preamble
struct RECORD_PREAMBLE {
    uint8_t  magic[6];        // "PONZU\0"
    uint8_t  record_type;     // 0 = SOA, 1 = file, etc.
    uint64_t data_len;        // # of blocks to read
    uint16_t modulo;          // # of bytes to use in last block
    uint8_t  checksum[64];    // SHA3-512 checksum.
}

*/

const (
	PREAMBLE_STRING = "PONZU"
)

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

type RecordType uint8
type RecordFlags uint16
type CompressionType uint8

const BLOCK_SIZE int64 = 4096

const (
	RECORD_TYPE_HEADER      RecordType = 0
	RECORD_TYPE_FILE        RecordType = 1
	RECORD_TYPE_HARDLINK    RecordType = 2
	RECORD_TYPE_SYMLINK     RecordType = 3
	RECORD_TYPE_DIRECTORY   RecordType = 4
	RECORD_TYPE_ZDICTIONARY RecordType = 5
	RECORD_TYPE_OS_SPECIAL  RecordType = 0x7F
)

const (
	RECORD_FLAG_NONE      RecordFlags = 0b00
	RECORD_FLAG_HALF      RecordFlags = 0b01
	RECORD_FLAG_STREAMED  RecordFlags = 0b10
	RECORD_FLAG_NO_CHKSUM RecordFlags = 0b10
)

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
)
