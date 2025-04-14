package base64

import (
	"encoding/base64"
	"io"

	"github.com/illikainen/go-utils/src/buffer"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Encoder struct {
	encoder *base64.Encoding
	writer  io.WriteSeeker
	width   int
	buffer  *buffer.Writer
}

var StdEncoding = base64.StdEncoding

func NewEncoder(enc *base64.Encoding, w io.WriteSeeker, width int) *Encoder {
	return &Encoder{
		encoder: enc,
		writer:  w,
		width:   width,
		buffer:  buffer.NewWriter(),
	}
}

func (e *Encoder) Write(p []byte) (int, error) {
	return e.buffer.Write(p)
}

func (e *Encoder) Seek(offset int64, whence int) (int64, error) {
	return e.buffer.Seek(offset, whence)
}

func (e *Encoder) Close() error {
	for _, chunk := range lo.Chunk(e.buffer.Bytes(), e.encoder.DecodedLen(e.width)) {
		data := []byte(e.encoder.EncodeToString(chunk) + "\n")
		n, err := e.writer.Write(data)
		if err != nil {
			return err
		}
		if n != len(data) {
			return errors.Errorf("invalid write")
		}
	}

	return nil
}
