package format

import (
	"bytes"

	"github.com/fxamacker/cbor/v2"
)

func UnmarshalRecordInfo(preamble *Preamble, data []byte) (*RecordInfo, error) {

	var info = &RecordInfo{}

	var err error = nil

	switch preamble.Rtype {
	case RECORD_TYPE_DIRECTORY:
		dir := &Directory{}
		err = cbor.Unmarshal(data, dir)
		info.Directory = dir
	case RECORD_TYPE_CONTROL:
		// handle SOA
		if preamble.Flags&RECORD_FLAG_CONTROL_START != 0 {
			soa := &StartOfArchive{}
			err = cbor.Unmarshal(data, soa)
			info.StartOfArchive = soa
		} else {
			unknown := &UnknownType{}
			err = cbor.Unmarshal(data, unknown.Info)
		}
	case RECORD_TYPE_FILE:
		file := &File{}
		err = cbor.Unmarshal(data, file)
		info.File = file
	case RECORD_TYPE_OS_SPECIAL:
		special := &OSSpecial{}
		err = cbor.Unmarshal(data, special)
		info.OSSpecial = special
	case RECORD_TYPE_CONTINUE:
		return nil, nil // Continue blocks never have metadata.
	case RECORD_TYPE_ZDICTIONARY:
		dict := &ZstdDictionary{}
		err = cbor.Unmarshal(data, dict)
		info.ZstdDictionary = dict
	case RECORD_TYPE_HARDLINK:
		link := &Hardlink{}
		err = cbor.Unmarshal(data, link)
		info.Hardlink = link
	case RECORD_TYPE_SYMLINK:
		link := &Symlink{}
		err = cbor.Unmarshal(data, link)
		info.Symlink = link
	default:
		dunno := &UnknownType{}
		err = cbor.Unmarshal(data, dunno.Info)
		if err != nil {
			return nil, err
		}
	}

	return info, err
}

func MarshalRecordInfo(rType RecordType, flags RecordFlags, info *RecordInfo) ([]byte, error) {
	buff := new(bytes.Buffer)
	encoder := cbor.NewEncoder(buff)

	var infoToEncode any
	switch rType {
	case RECORD_TYPE_CONTINUE:
		// :) Continue has no fields.
	case RECORD_TYPE_CONTROL:
		if flags == RECORD_FLAG_CONTROL_START {
			infoToEncode = info.StartOfArchive
		} else {
			return nil, nil
		}
	case RECORD_TYPE_FILE:
		infoToEncode = info.File
	case RECORD_TYPE_DIRECTORY:
		infoToEncode = info.Directory
	case RECORD_TYPE_ZDICTIONARY:
		infoToEncode = info.ZstdDictionary
	case RECORD_TYPE_HARDLINK:
		infoToEncode = info.Hardlink
	case RECORD_TYPE_SYMLINK:
		infoToEncode = info.Symlink
	case RECORD_TYPE_OS_SPECIAL:
		infoToEncode = info.OSSpecial
	default:
		// Don't know, try anyway
		infoToEncode = info
	}

	err := encoder.Encode(infoToEncode)
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
