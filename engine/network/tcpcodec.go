package network

import (
	"bufio"
	"io"
)

type TCPCodec interface {
	Read(r io.Reader) ([]byte, error)
	Write(w io.Writer, b []byte) error
}

type defaultTCPCodec struct {
}

var DefaultTCPCodec = &defaultTCPCodec{}

func (this *defaultTCPCodec) Read(r io.Reader) ([]byte, error) {
	rBuf := bufio.NewReader(r)
	if _, err := rBuf.Peek(1); err != nil {
		return nil, err
	}
	b := make([]byte, rBuf.Buffered())
	if _, err := rBuf.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (this *defaultTCPCodec) Write(w io.Writer, b []byte) error {
	if _, err := w.Write(b); err != nil {
		return err
	}
	return nil
}
