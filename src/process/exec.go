package process

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/illikainen/go-utils/src/errorx"
	"github.com/illikainen/go-utils/src/logging"
	"github.com/illikainen/go-utils/src/stringx"
)

type OutputFunc = func(io.Reader, int) ([]byte, error)

type ExecOptions struct {
	Command []string
	Env     []string
	Become  string
	Stdin   io.Reader
	Stdout  OutputFunc
	Stderr  OutputFunc
}

const (
	Stdout = iota
	Stderr
)

type ExecOutput struct {
	Stdout []byte
	Stderr []byte
}

func Exec(opts *ExecOptions) (*ExecOutput, error) {
	var cmd *exec.Cmd
	if opts.Become != "" {
		esc, err := become(opts.Become)
		if err != nil {
			return nil, err
		}

		cmd = exec.Command(esc[0], append(esc[1:], opts.Command...)...) // #nosec G204
	} else {
		cmd = exec.Command(opts.Command[0], opts.Command[1:]...) // #nosec G204
	}

	cmd.Env = opts.Env
	cmd.Stdin = opts.Stdin

	out := &ExecOutput{}
	group := errgroup.Group{}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stdoutFunc := opts.Stdout
	if stdoutFunc == nil {
		stdoutFunc = CaptureOutput
	}

	group.Go(func() error {
		out.Stdout, err = stdoutFunc(stdoutPipe, Stdout)
		return err
	})

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stderrFunc := opts.Stderr
	if stderrFunc == nil {
		stderrFunc = CaptureOutput
	}

	group.Go(func() error {
		out.Stderr, err = stderrFunc(stderrPipe, Stderr)
		return err
	})

	log.Tracef("exec: %s", strings.Join(cmd.Args, " "))
	err = cmd.Start()
	if err != nil {
		// XXX: what to do with the goroutine errgroup here?
		return nil, err
	}

	err = group.Wait()
	if err != nil {
		return nil, errorx.Join(err, cmd.Wait())
	}

	err = cmd.Wait()
	if err != nil {
		if len(out.Stderr) > 0 {
			return nil, errors.Errorf("%s", out.Stderr)
		}
		return nil, err
	}

	return out, nil
}

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
			return nil, err
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

func become(username string) ([]string, error) {
	cur, err := user.Current()
	if err != nil {
		return nil, err
	}

	if cur.Username == username {
		return []string{}, nil
	}

	sudo, err := exec.LookPath("sudo")
	if err == nil {
		return []string{sudo, "-u", username}, nil
	}

	doas, err := exec.LookPath("doas")
	if err == nil {
		return []string{doas, "-u", username}, nil
	}

	return nil, errors.Errorf("unable to find a suitable program to change privileges")
}
