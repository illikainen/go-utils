package process

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/illikainen/go-utils/src/logging"
	"github.com/illikainen/go-utils/src/stringx"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type OutputFunc = func(io.Reader, int, bool) ([]byte, error)

// FIXME: do something about the blocking behavior when io.Copy()ing from a
// pipe to a *bytes.Buffer in ByteOutput() so that we don't need this function
// when shipping data from one process to another.
func UnsafeByteOutput(reader io.Reader, src int, _ bool) ([]byte, error) {
	w := os.Stdout
	if src != Stdout {
		w = os.Stderr
	}

	_, err := io.Copy(w, reader)
	return nil, err
}

func ByteOutput(reader io.Reader, src int, trusted bool) ([]byte, error) {
	buf := &bytes.Buffer{}
	n, err := io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}

	if !trusted && !bytes.Equal(stringx.Sanitize(buf.Bytes()), buf.Bytes()) {
		return nil, errors.Errorf("ByteOutput(): invalid data")
	}

	w := os.Stdout
	if src != Stdout {
		w = os.Stderr
	}

	m, err := io.Copy(w, buf)
	if err != nil {
		return nil, err
	}
	if n != m || m < math.MinInt || m > math.MaxInt {
		return nil, errors.Errorf("ByteOutput(): invalid data size")
	}

	return buf.Bytes(), nil
}

func CaptureOutput(reader io.Reader, _ int, trusted bool) ([]byte, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanBytes)

	data := []byte{}
	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	if !trusted && !bytes.Equal(stringx.Sanitize(data), data) {
		return nil, errors.Errorf("CaptureOutput(): invalid data")
	}

	return data, nil
}

func TextOutput(reader io.Reader, src int, trusted bool) ([]byte, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	w := os.Stdout
	if src != Stdout {
		w = os.Stderr
	}

	data := []byte{}
	for scanner.Scan() {
		chunk := scanner.Bytes()
		if !trusted && !bytes.Equal(stringx.Sanitize(chunk), chunk) {
			return nil, errors.Errorf("TextOutput(): invalid data")
		}

		str := fmt.Sprintf("%s\n", chunk)
		n, err := fmt.Fprintf(w, "%s", str)
		if err != nil {
			return nil, err
		}

		if n != len(str) {
			return nil, errors.Errorf("unexpected write, %d != %d", n, len(str))
		}

		data = append(data, chunk...)
		data = append(data, '\n')
	}

	return data, scanner.Err()
}

func LogrusOutput(reader io.Reader, _ int, trusted bool) ([]byte, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	data := []byte{}
	for scanner.Scan() {
		chunk := scanner.Bytes()
		if !trusted && !bytes.Equal(stringx.Sanitize(chunk), chunk) {
			return nil, errors.Errorf("TextOutput(): invalid data")
		}

		var fields log.Fields
		err := json.Unmarshal(chunk, &fields)
		if err != nil {
			fields = log.Fields{}
			fields["msg"] = string(chunk)
			fields["unstyled"] = true
		}

		level, err := log.ParseLevel(logging.GetField(fields, "level", "info"))
		if err != nil {
			return nil, err
		}

		msg := logging.GetField(fields, "msg", "n/a")

		// The reason we don't log fatal messages is that they result in a
		// non-zero exit code.  After Cmd.Wait() fails because of the exit
		// code, the fatal message is propagated to the caller through the
		// output buffer.  If we'd log them here, we'd end up with
		// duplicate fatal messages.
		if level != log.FatalLevel {
			log.WithFields(fields).Logln(level, msg)
		}

		// The reason we clear the buffer is that we don't want stderr
		// polluted with valid data (e.g., normal log messages) when we
		// use the error message in the parent process.
		if level == log.FatalLevel {
			data = []byte{}
		}

		data = append(data, msg...)
		data = append(data, '\n')
	}

	return data, scanner.Err()
}
