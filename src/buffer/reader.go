package buffer

import (
	"io"
	"math"
	"os"
	"time"

	"github.com/pkg/errors"
)

type Reader struct {
	data     []byte
	position int
}

func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

func (r *Reader) Read(p []byte) (int, error) {
	if r.position >= len(r.data) {
		return 0, io.EOF
	}

	n := copy(p, r.data[r.position:])
	if r.position > math.MaxInt-n {
		return 0, os.ErrInvalid
	}
	r.position += n

	return n, nil
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	if offset < math.MinInt || offset > math.MaxInt-int64(len(r.data)) {
		return 0, errors.Errorf("invalid offset: %d", offset)
	}

	switch whence {
	case io.SeekStart:
		r.position = int(offset)
	case io.SeekCurrent:
		r.position += int(offset)
	case io.SeekEnd:
		r.position = len(r.data) + int(offset)
	default:
		return 0, errors.Errorf("invalid whence: %d", whence)
	}

	//lint:ignore SA4003 sanity
	if r.position < 0 || r.position < math.MinInt64 || r.position > math.MaxInt64 {
		return 0, errors.Errorf("invalid position: %d", r.position)
	}
	return int64(r.position), nil
}

func (r *Reader) Sync() error {
	return nil
}

func (r *Reader) Stat() (os.FileInfo, error) {
	size := len(r.data)
	//lint:ignore SA4003 sanity
	if size < math.MinInt64 || size > math.MaxInt64 {
		return nil, errors.Errorf("invalid size")
	}
	return &FileInfo{size: int64(size)}, nil
}

type FileInfo struct {
	size int64
}

func (fi *FileInfo) Name() string {
	return "buffer"
}

func (fi *FileInfo) Size() int64 {
	return fi.size
}

func (fi *FileInfo) Mode() os.FileMode {
	return 0
}

func (fi *FileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *FileInfo) IsDir() bool {
	return false
}

func (fi *FileInfo) Sys() any {
	return nil
}
