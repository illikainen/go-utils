package base64

import (
	"encoding/base64"
	"io"
	"os"

	"github.com/illikainen/go-utils/src/buffer"
)

type Decoder struct {
	buffer *buffer.Reader
}

func NewDecoder(enc *base64.Encoding, r io.ReadSeeker) (*Decoder, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	decoded, err := enc.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	return &Decoder{
		buffer: buffer.NewReader(decoded),
	}, nil
}

func (d *Decoder) Read(p []byte) (int, error) {
	return d.buffer.Read(p)
}

func (d *Decoder) Seek(offset int64, whence int) (int64, error) {
	return d.buffer.Seek(offset, whence)
}

func (d *Decoder) Stat() (os.FileInfo, error) {
	return d.buffer.Stat()
}

func (d *Decoder) Sync() error {
	return d.buffer.Sync()
}

func (d *Decoder) Name() string {
	return "base64decoder"
}
