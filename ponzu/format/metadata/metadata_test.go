package metadata

import (
	"bytes"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

// This test is here to make sure I understand the behavior of pointers in cbor's interpretation.
// Since all the particulars
func TestCommon(t *testing.T) {

	commonMeta := new(CommonMetadata)
	commonMeta.FileSize = MakePointer[uint64](5)

	buff := new(bytes.Buffer)

	enc := cbor.NewEncoder(buff)
	if nil != enc.Encode(commonMeta) {
		t.Fatal("Couldn't encode!")
	}
	b := buff.Bytes()
	buff.Reset()
	buff.Write(b)
	dec := cbor.NewDecoder(buff)

	newMeta := new(CommonMetadata)
	if nil != dec.Decode(newMeta) {
		t.Fatal("Couldn't decode!")
	}

	if *(newMeta.FileSize) != 5 {
		t.Fatal("Didn't get the same value in that I got out")
	}
	if newMeta.FileSize == commonMeta.FileSize {
		t.Fatal("Shouldn't be the same address...")
	}

	t.Log(newMeta.FileSize)
	t.Log(commonMeta.FileSize)
}
