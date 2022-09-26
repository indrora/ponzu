package writer

import (
	"bytes"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/indrora/ponzu/ponzu/format"
)

func TestWriter(t *testing.T) {

	buffer := new(bytes.Buffer)

	writer := NewWriter(buffer, 100)

	data := []byte{1, 2, 3, 4}

	metadata := map[string]string{
		"hello": "world",
	}

	if writer.AppendSOA("test", "bueno") != nil {
		t.Error("Failed to append SOA!")
		t.Fail()
	}

	if writer.AppendBytes(254, 0, metadata, data) != nil {
		t.Error("Failed to append some bytes....")
		t.Fail()
	}

	if writer.Close() != nil {
		t.Error("Failed to close the archive...")
		t.Fail()
	}

	if int64(buffer.Len()) != 3*format.BLOCK_SIZE {
		t.Errorf("Expected more bytes. Got %d, expected %d", buffer.Len(), 3*format.BLOCK_SIZE)
		t.Log(spew.Sdump(buffer.Bytes()))

		t.Fail()
	}

}
