package process

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/illikainen/go-utils/src/logging"
	"github.com/illikainen/go-utils/src/stringx"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type OutputFunc = func(io.Reader, int) ([]byte, error)

// NOTE: neither the printed output nor the output buffer is sanitized
func ByteOutput(reader io.Reader, src int) ([]byte, error) {
	w := os.Stdout
	if src != Stdout {
		w = os.Stderr
	}

	_, err := io.Copy(w, reader)
	return nil, err
}

// NOTE: the output buffer isn't sanitized
func CaptureOutput(reader io.Reader, _ int) ([]byte, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanBytes)

	data := []byte{}
	for scanner.Scan() {
		data = append(data, scanner.Bytes()...)
	}

	return data, scanner.Err()
}

// NOTE: the printed output is sanitized but the output buffer is not
func TextOutput(reader io.Reader, src int) ([]byte, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	data := []byte{}
	for scanner.Scan() {
		w := os.Stdout
		if src != Stdout {
			w = os.Stderr
		}

		str := fmt.Sprintf("%s\n", stringx.Sanitize(scanner.Text()))
		n, err := fmt.Fprintf(w, "%s", str)
		if err != nil {
			return nil, err
		}

		if n != len(str) {
			return nil, errors.Errorf("unexpected write, %d != %d", n, len(str))
		}

		data = append(data, scanner.Bytes()...)
		data = append(data, '\n')
	}

	return data, scanner.Err()
}

// NOTE: the printed output is sanitized but the output buffer is not
func LogrusOutput(reader io.Reader, _ int) ([]byte, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	data := []byte{}
	for scanner.Scan() {
		var fields log.Fields
		err := json.Unmarshal(stringx.Sanitize(scanner.Bytes()), &fields)
		if err != nil {
			fields = log.Fields{}
			fields["msg"] = stringx.Sanitize(scanner.Text())
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
