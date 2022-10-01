package loopreader

import (
	"bytes"
	"io"
)

type Reader struct {
	data   []byte
	reader io.Reader
}

func New(reader io.Reader) (*Reader, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return &Reader{
		data:   b,
		reader: bytes.NewBuffer(b),
	}, nil
}

func (t *Reader) Close() error { return nil }

func (t *Reader) Read(in []byte) (n int, err error) {
	n, err = t.reader.Read(in)
	if err == io.EOF {
		t.reader = bytes.NewReader(t.data)
	}
	return n, err
}
