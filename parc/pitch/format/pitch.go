package format

type PTimestamp uint64

// 70-byte preamble
// This preamble can be used to read the type of record you've just recieved.
type PitchPreamble struct {
	Magic    [5]byte
	Rtype    RecordType
	Flags    RecordFlags // flag set
	Checksum [64]byte
}

// Header
type PitchArchiveHeader struct {
	// Preamble (included here for simplicity)
	PitchPreamble
	// Version of the archive format to reference
	Version uint16
	// Date of archive creation (0 is a valid value)
	Created PTimestamp
	// Host OS that is used to create this archive
	HostOS [24]byte
	// Compression used in this archive
	Compression CompressionType
	// How many 4K blocks this archive takes up
	BlockCount uint64
	// Padding to the back half of the header.
	padding [1943]byte // size = 2048-(70+2+4+24+1+4)
	// Filename prefix used by all items within this archive
	Prefix [1024]byte
	// String comment
	Comment [1024]byte
}

type PitchFileHeader struct {
	// 70 byte preamble
	PitchPreamble
	// Path of the file
	Path [1024]byte
	// How many full blocks does this take up
	BlockCount uint64
	// size mod blocksize
	Modulo uint16
	// unix permissions of the inode
	Permissions uint16
	// Modification time of the file
	Modtime PTimestamp
	// padding to take us to the end of the file header
	padding [942]byte // size = 2048-(70+1024+4+2+2+4)
	// OS-specific attributes
	Attributes [1024]byte
}

// A short file record
type PitchMiniFile struct {
	// 70-byte preamble
	PitchPreamble
	// Path to the file
	Path [768]byte
	// Size of the file in bytes
	Size uint16
	// UNIX permissions of the file
	Permissions uint16
	Modtime     PTimestamp
	Attributes  [1024]byte
	padding     [178]byte // size = 2048 - (70+768+2+2+4+1024)
	Data        [2048]byte
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
	RECORD_TYPE_OS_SPECIAL  RecordType = 0xFF
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
