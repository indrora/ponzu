package writer

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"
	"time"

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

// Test to make sure that the whole thing gets read properly.
func TestWriterEncode(t *testing.T) {

	// I don't *need* to add a start of archive, so we're going to make up some bullshit

	buff := new(bytes.Buffer)

	writer := NewWriter(buff, 0)

	// Generate some data

	randData := make([]byte, int(1.50*float32(format.BLOCK_SIZE)))
	rand.Read(randData)
	fileinfo := format.File{
		Name:       "foo",
		Mode:       0666,
		Owner:      "billy",
		Group:      "billy",
		Compressor: format.COMPRESSION_NONE,
		ModTime:    time.Now(),
	}
	writer.AppendBytes(format.RECORD_TYPE_FILE, format.RECORD_FLAG_NONE, fileinfo, randData)
	writer.Close()

	newbytes := buff.Bytes()
	if int64(len(newbytes)) != 3*format.BLOCK_SIZE {
		t.Errorf("Expected %d bytes, got %d", 3*format.BLOCK_SIZE, len(newbytes))
	}

	// Check that the length and such are properly encoded
	if newbytes[6] != byte(format.RECORD_TYPE_FILE) {
		t.Errorf("Expected FILE, got %d", newbytes[6])
	}

	if testFlags := binary.BigEndian.Uint16(newbytes[7:]); testFlags != 0 {
		t.Errorf("Expected 0 flags, got %d", testFlags)
	}

	if testBcount := binary.BigEndian.Uint64(newbytes[9:]); testBcount != 2 {
		t.Errorf("Expected 2 length, got %d", testBcount)
	}

	checkModulo := uint16(len(randData) % int(format.BLOCK_SIZE))
	if testModulo := binary.BigEndian.Uint16(newbytes[17:]); testModulo != checkModulo {
		t.Errorf("Expected modulous to be %d, got %d", checkModulo, testModulo)
	}

	spew.Dump(newbytes[:70])

	// We should be able to trim off the first block and be able to read the same data we put in back, but padded out

	if newbytes[4096] != randData[0] {
		t.Error("First byte of data block is not correct data")
	}

	t.Logf("data len = %d", len(randData))

	for i, v := range newbytes[4096:] {
		if i >= len(randData) {
			if v != 0 {
				t.Errorf("Wrong data in padding, expecting 0, got %d", v)
			}
		} else if randData[i] != v {
			t.Errorf("Bad data at %v; expected %v, got %v", i, randData[i], v)
		}
	}
}

func TestChecksum(t *testing.T) {

}
