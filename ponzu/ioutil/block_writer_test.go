package ioutil

import (
	"bytes"
	"crypto/rand"
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

func TestMultiBlock(t *testing.T) {
	buffer := new(bytes.Buffer)
	writer := NewBlockWriter(buffer, 16)

	// we're going to write some big buckets of data

	some_data := make([]byte, 27)
	rand.Read(some_data)

	writer.Write(some_data)
	t.Log(spew.Sprint(buffer.Bytes()))
	// Check the internal bit
	if writer.writtenSinceRealign != 27 {
		t.Fail()
	}
	// we haven't aligned the write yet, so
	if buffer.Len() != len(some_data) {
		t.Fail()
	}
	// forward to the next block
	writer.Align()
	t.Log(spew.Sprint(buffer.Bytes()))
	if buffer.Len() != 32 {
		t.Fail()
	}

}
