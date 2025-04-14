package buffer

import (
	"io"
	"math"

	"github.com/pkg/errors"
)

type Writer struct {
	data     []byte
	position int
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Write(p []byte) (int, error) {
	size := len(w.data)
	if w.position > size {
		chunk := make([]byte, w.position-size)
		w.data = append(w.data, chunk...)
	}

	if w.position == size {
		w.data = append(w.data, p...)
	} else {
		for i := 0; i < len(p); i++ {
			if w.position+i < len(w.data) {
				w.data[w.position+i] = p[i]
			} else {
				w.data = append(w.data, p[i])
			}
		}
	}

	w.position += len(p)
	return len(p), nil
}

func (w *Writer) Seek(offset int64, whence int) (int64, error) {
	if offset < math.MinInt || offset > math.MaxInt-int64(len(w.data)) {
		return 0, errors.Errorf("invalid offset: %d", offset)
	}

	switch whence {
	case io.SeekStart:
		w.position = int(offset)
	case io.SeekCurrent:
		w.position += int(offset)
	case io.SeekEnd:
		w.position = len(w.data) + int(offset)
	default:
		return 0, errors.Errorf("invalid whence: %d", whence)
	}

	//lint:ignore SA4003 sanity
	if w.position < 0 || w.position < math.MinInt64 || w.position > math.MaxInt64 {
		return 0, errors.Errorf("invalid position: %d", w.position)
	}
	return int64(w.position), nil
}

func (w *Writer) Bytes() []byte {
	return w.data
}
