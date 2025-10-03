package reader

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/indrora/ponzu/ponzu/format"
)

func UnmarshalRecordInfo(preamble *format.Preamble, data []byte) (*format.RecordInfo, error) {

	var info = &format.RecordInfo{}

	var err error = nil

	fmt.Printf("rtype = %d size = %d\n", preamble.Rtype, preamble.InfoLength)

	switch preamble.Rtype {
	case format.RECORD_TYPE_DIRECTORY:
		dir := &format.Directory{}
		err = cbor.Unmarshal(data, dir)
		info.Directory = dir
	case format.RECORD_TYPE_CONTROL:
		// handle SOA
		if preamble.Flags&format.RECORD_FLAG_CONTROL_START != 0 {
			soa := &format.StartOfArchive{}
			err = cbor.Unmarshal(data, soa)
			info.StartOfArchive = soa
		} else {
			unknown := &format.UnknownType{}
			err = cbor.Unmarshal(data, unknown)
			info.UnknownType = unknown
		}
	case format.RECORD_TYPE_FILE:
		file := &format.File{}
		err = cbor.Unmarshal(data, file)
		info.File = file
	case format.RECORD_TYPE_OS_SPECIAL:
		special := &format.OSSpecial{}
		err = cbor.Unmarshal(data, special)
		info.OSSpecial = special
	case format.RECORD_TYPE_CONTINUE:
		return nil, nil // Continue blocks never have metadata.
	case format.RECORD_TYPE_ZDICTIONARY:
		dict := &format.ZstdDictionary{}
		err = cbor.Unmarshal(data, dict)
		info.ZstdDictionary = dict
	case format.RECORD_TYPE_HARDLINK:
		link := &format.Hardlink{}
		err = cbor.Unmarshal(data, link)
		info.Hardlink = link
	case format.RECORD_TYPE_SYMLINK:
		link := &format.Symlink{}
		err = cbor.Unmarshal(data, link)
		info.Symlink = link
	default:
		dunno := &format.UnknownType{}
		err = cbor.Unmarshal(data, dunno)
		if err != nil {
			return nil, err
		}
		info.UnknownType = dunno
	}

	return info, err
}
