package ioutil

import (
	"bytes"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestBlockWriterSimple(t *testing.T) {

	buffer := new(bytes.Buffer)

	writer := NewBlockWriter(buffer, 100)
	if _, err := writer.Write([]byte{1, 2, 3, 4}); err != nil {
		t.Fail()
	}

	if len(buffer.Bytes()) > 4 {
		t.Error("Too many bytes written before align")
	}

	if writer.Align() != nil {
		t.Error("Failed to align to end of buffer")
	}

	buflen := len(buffer.Bytes())
	if len(buffer.Bytes()) != 100 {
		t.Errorf("Expected 100 bytes written, got %d", buflen)
	}

	if writer.Align() != nil {
		t.Errorf("should be able to align twice.")
	}
	if len(buffer.Bytes()) != buflen {
		t.Errorf("Repeated aligns should not have any effect.")
	}

	t.Logf(spew.Sprint(buffer.Bytes()))

}
