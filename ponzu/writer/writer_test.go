package writer

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/indrora/ponzu/ponzu/format"
	"github.com/indrora/ponzu/ponzu/ioutil"
)

func TestWriter(t *testing.T) {

	buffer := new(bytes.Buffer)

	writer := NewWriter(buffer, 100)

	data := []byte{1, 2, 3, 4}

	metadata := map[string]string{
		"hello": "world",
	}

	if writer.AppendStart("test", "bueno") != nil {
		t.Error("Failed to append SOA!")
		t.Fail()
	}

	if writer.AppendBytes(254, 0, format.COMPRESSION_NONE, metadata, data) != nil {
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

	randData := make([]byte, int(1.75*float32(format.BLOCK_SIZE)))
	rand.Read(randData)
	fileinfo := format.File{
		Name:    "foo",
		ModTime: time.Now(),
	}

	test_rType := uint8(rand.Intn(255))
	test_rFlags := uint16(rand.Intn(255))

	err := writer.AppendBytes(format.RecordType(test_rType), format.RecordFlags(test_rFlags), format.COMPRESSION_NONE, fileinfo, randData)

	if err != nil {
		t.Fatal(err, "Failed to write to the archive!")
	}

	writer.Close()

	newbytes := buff.Bytes()

	if int64(len(newbytes)) != 3*format.BLOCK_SIZE {
		t.Fatalf("Wrong number of bytes written: Expected %v, got %v", 3*format.BLOCK_SIZE, len(newbytes))
	}

	bufread := bytes.NewReader(newbytes)

	blockread := ioutil.NewBlockReader(bufread, format.BLOCK_SIZE)

	fileheader, _ := blockread.ReadBlock()
	fileheaderReader := bytes.NewReader(fileheader)

	testPreamble, err := format.ReadPreamble(fileheaderReader)
	if err != nil {
		t.Errorf("Failed to read preamble: %v", err)
	}

	if testPreamble.Rtype != format.RecordType(test_rType) {
		t.Errorf("Expected record type %v, got %v", test_rType, testPreamble.Rtype)
	}
	if testPreamble.Flags != format.RecordFlags(test_rFlags) {
		t.Errorf("Expected record flag %v, got %v", test_rFlags, testPreamble.Rtype)
	}

	if testPreamble.DataLen != 2 {
		t.Errorf("Expected 2 length, got %d", testPreamble.DataLen)
	}

	checkModulo := uint16(len(randData) % int(format.BLOCK_SIZE))
	if testPreamble.Modulo != checkModulo {
		t.Errorf("Expected modulous to be %d, got %d", checkModulo, testPreamble.Modulo)
	}

	// did we also get data right?

	testData, err := blockread.ReadBlock()
	if err == io.EOF {
		t.Error(err, "Unexpected EOF")
	}
	// now, we should be able to read the next block and get an EOF
	testData2, err := blockread.ReadBlock()
	if err != io.EOF {
		remain, e := blockread.ReadBlock()
		spew.Dump(testData)
		t.Error(err, e, "expected EOF, got something else")
	}
	testData = append(testData, testData2[:testPreamble.Modulo]...)
	if !bytes.Equal(testData, randData) {
		t.Error("failed to read the appropriate data size.")
	}

}
