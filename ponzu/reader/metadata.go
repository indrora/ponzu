package reader

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/indrora/ponzu/ponzu/format"
)

func unmarshalMetadata(preamble *format.Preamble, data []byte) any {

	switch preamble.Rtype {
	case format.RECORD_TYPE_CONTROL:
		if preamble.Flags == format.RECORD_FLAG_CONTROL_START {
			return unmarshalOrNil[format.StartOfArchive](data)
		}
	case format.RECORD_TYPE_DIRECTORY:
		return unmarshalOrNil[format.Directory](data)
	case format.RECORD_TYPE_FILE:
		return unmarshalOrNil[format.File](data)
	case format.RECORD_TYPE_CONTINUE:
		return nil // Continue blocks never have metadata.
	case format.RECORD_TYPE_OS_SPECIAL:
		return unmarshalOrNil[format.OSSpecial](data)
	default:
		return nil
	}

	return nil
}

func unmarshalOrNil[T any](data []byte) *T {
	ret := new(T)
	if err := cbor.Unmarshal(data, ret); err == nil {

		return ret
	} else {
		fmt.Println("xxx: couldn't unmarshal: ", err)
	}
	return nil

}
