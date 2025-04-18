package iofs

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/illikainen/go-utils/src/errorx"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type FileInfo interface {
	Name() string
	Stat() (os.FileInfo, error)
}

var ErrInvalidSize = errors.New("invalid size")
var ErrInvalidOffset = errors.New("invalid offset")

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func MoveFile(src string, dst string) error {
	err := Copy(dst, src)
	if err != nil {
		return err
	}

	err = os.Remove(src)
	if err != nil {
		return err
	}

	return nil
}

func ReadFile(path string) ([]byte, error) {
	buf := bytes.Buffer{}
	err := Copy(&buf, path)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Copy[T any, U any](dst T, src U) (err error) {
	var dstf io.Writer
	dstName := ""

	switch dst := any(dst).(type) {
	case io.Writer:
		fi, ok := any(dst).(FileInfo)
		if ok {
			dstName = fi.Name()
		} else {
			dstName = fmt.Sprintf("%p", dst)
		}
		dstf = dst
	case string:
		dir, _ := filepath.Split(dst)
		if dir != "" {
			err := os.MkdirAll(dir, 0700)
			if err != nil {
				return err
			}
		}

		f, err := os.Create(dst) // #nosec G304
		if err != nil {
			return err
		}
		defer errorx.Defer(f.Close, &err)
		dstf = f
		dstName = dst
	default:
		return errors.Errorf("invalid dst")
	}

	var srcf io.Reader
	srcName := ""
	srcSize := int64(-1)

	switch src := any(src).(type) {
	case io.Reader:
		fi, ok := any(src).(FileInfo)
		if ok {
			s, err := fi.Stat()
			if err != nil {
				return err
			}

			srcSize = s.Size()
			if srcSize < 0 {
				return ErrInvalidSize
			}

			srcName = fi.Name()
		} else {
			srcName = fmt.Sprintf("%p", src)
		}

		seeker, ok := any(src).(io.Seeker)
		if ok && srcSize > 0 {
			n, err := seeker.Seek(0, io.SeekCurrent)
			if err != nil {
				return err
			}
			srcSize -= n
		}

		srcf = src
	case *zip.File:
		f, err := src.Open()
		if err != nil {
			return err
		}
		defer errorx.Defer(f.Close, &err)

		info := src.FileInfo()
		srcSize = info.Size()
		srcName = info.Name()
		srcf = f
	case string:
		f, err := os.Open(src) // #nosec G304
		if err != nil {
			return err
		}
		defer errorx.Defer(f.Close, &err)

		stat, err := f.Stat()
		if err != nil {
			return err
		}

		srcSize = stat.Size()
		srcName = src
		srcf = f
	default:
		return errors.Errorf("invalid src")
	}

	log.Tracef("%s: copy %d byte(s) from %s", dstName, srcSize, srcName)
	n, err := io.Copy(dstf, srcf)
	if err != nil {
		return err
	}

	if srcSize >= 0 && n != srcSize {
		return errors.Wrap(ErrInvalidSize, srcName)
	}

	return nil
}

func ReadFull(r io.Reader, buf []byte) error {
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return ErrInvalidSize
	}
	return nil
}

func MkdirTemp() (string, func() error, error) {
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return "", nil, err
	}

	return tmp, func() error { return os.RemoveAll(tmp) }, nil
}

func Expand(path string) (string, error) {
	if path == "" {
		return "", errors.Errorf("empty path")
	}

	sep := string(os.PathSeparator)

	if strings.HasPrefix(path, "~"+sep) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		path = filepath.Join(home, path[1+len(sep):])
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return path, nil
}

func Seek(s io.Seeker, offset int64, whence int) (int64, error) {
	n, err := s.Seek(offset, whence)
	if err != nil {
		return 0, err
	}

	if n != offset {
		return 0, ErrInvalidOffset
	}

	return n, nil
}

func Remove(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if stat.IsDir() {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}
