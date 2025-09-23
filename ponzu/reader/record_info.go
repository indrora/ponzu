package reader

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/indrora/ponzu/ponzu/format"
)

func UnmarshalRecordInfo(preamble *format.Preamble, data []byte) (*format.RecordInfo, error) {

	var info = &format.RecordInfo{}

	var err error = nil

	switch preamble.Rtype {
	case format.RECORD_TYPE_DIRECTORY:
		err = cbor.Unmarshal(data, info.Directory)
	case format.RECORD_TYPE_CONTROL:
		// handle SOA
		if preamble.Flags&format.RECORD_FLAG_CONTROL_START != 0 {
			err = cbor.Unmarshal(data, info.StartOfArchive)
		} else {
			err = cbor.Unmarshal(data, info.UnknownType)
		}
	case format.RECORD_TYPE_FILE:
		err = cbor.Unmarshal(data, info.File)
	case format.RECORD_TYPE_OS_SPECIAL:
		err = cbor.Unmarshal(data, info.OSSpecial)
	case format.RECORD_TYPE_CONTINUE:
		return nil, nil // Continue blocks never have metadata.
	case format.RECORD_TYPE_ZDICTIONARY:
		err = cbor.Unmarshal(data, info.ZstdDictionary)
	case format.RECORD_TYPE_HARDLINK:
		err = cbor.Unmarshal(data, info.Hardlink)
	case format.RECORD_TYPE_SYMLINK:
		err = cbor.Unmarshal(data, info.Symlink)
	default:
		err = cbor.Unmarshal(data, info.UnknownType)
		if err != nil {
			return nil, err
		}
	}

	return info, err
}
