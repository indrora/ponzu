package ioutil

import (
	"bytes"
	"io"
)

type Pump struct {
	source io.Reader
	pipe   io.ReadWriter
}

func NewPump(reader io.Reader, pipe io.ReadWriter) *Pump {
	return &Pump{
		source: reader,
		pipe:   pipe,
	}
}

func (pump *Pump) Read(buf []byte) (int64, error) {
	intBuf := bytes.NewBuffer(buf)
	// Copy from the source -> internal buffer
	_, err1 := io.Copy(intBuf, pump.source)
	if err1 != nil && err1 != io.EOF {
		return 0, err1
	}
	// Copy from the internal buffer -> pipe
	io.Copy(pump.pipe, intBuf)
	// copy from the pipe -> buffer
	ib, _ := io.Copy(intBuf, pump.pipe)

	return ib, err1
}
