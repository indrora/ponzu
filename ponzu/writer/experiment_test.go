package writer_test

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestExperiment(t *testing.T) {

	st, e := os.Stat("experiment_test.go")
	if e != nil {
		t.FailNow()
	}
	f, e := os.Open("experiment_test.go")

	if e != nil {
		t.Error("Failed to open myself")
	}

	readTotal := int64(0)
	eachRead := st.Size() / 2
	defer f.Close()

	buff := new(bytes.Buffer)

	t.Logf("File is %v long", st.Size())

	for e == nil {
		r := int64(0)
		r, e = io.CopyN(buff, f, eachRead)
		if e == io.EOF {
			t.Logf("Found EOF, read %v bytes", r)
		} else {
			t.Logf("Did not get EOF, read %v", r)
		}
		readTotal += r
		//...  asdf
	}

}
