package process

import (
	"io"
	"os/exec"
	"os/user"
	"strings"

	"github.com/illikainen/go-utils/src/errorx"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type ExecOptions struct {
	Command         []string
	Env             []string
	Dir             string
	Become          string
	Stdin           io.Reader
	Stdout          OutputFunc
	Stderr          OutputFunc
	IgnoreExitError bool
	Trusted         bool
}

const (
	Stdout = iota
	Stderr
)

type ExecOutput struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

func Exec(opts *ExecOptions) (*ExecOutput, error) {
	args := opts.Command
	if opts.Become != "" {
		esc, err := Become(opts.Become)
		if err != nil {
			return nil, err
		}

		args = append(esc, args...)
	}

	cmd := exec.Command(args[0], args[1:]...) // #nosec G204
	cmd.Env = opts.Env
	cmd.Dir = opts.Dir
	cmd.Stdin = opts.Stdin

	out := &ExecOutput{ExitCode: 1}
	group := errgroup.Group{}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	stdoutFunc := opts.Stdout
	if stdoutFunc == nil {
		stdoutFunc = CaptureOutput
	}

	group.Go(func() error {
		out.Stdout, err = stdoutFunc(stdoutPipe, Stdout, opts.Trusted)
		return err
	})

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	stderrFunc := opts.Stderr
	if stderrFunc == nil {
		stderrFunc = CaptureOutput
	}

	group.Go(func() error {
		out.Stderr, err = stderrFunc(stderrPipe, Stderr, opts.Trusted)
		return err
	})

	log.Tracef("exec: %s", strings.Join(cmd.Args, " "))
	err = cmd.Start()
	if err != nil {
		// XXX: what to do with the goroutine errgroup here?
		return nil, errors.WithStack(err)
	}

	err = group.Wait()
	if err != nil {
		return nil, errorx.Join(errors.WithStack(err), cmd.Wait())
	}

	err = cmd.Wait()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			out.ExitCode = exitErr.ExitCode()

			if opts.IgnoreExitError {
				return out, nil
			}

			if len(out.Stderr) > 0 {
				return nil, errors.Errorf("%s", out.Stderr)
			}
		}
		return nil, errors.WithStack(err)
	}

	out.ExitCode = 0
	return out, nil
}

func Become(username string) ([]string, error) {
	cur, err := user.Current()
	if err != nil {
		return nil, errors.WithStack(err)
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
