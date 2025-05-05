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
	Command []string
	Env     []string
	Dir     string
	Become  string
	Stdin   io.Reader
	Stdout  OutputFunc
	Stderr  OutputFunc
	Trusted bool
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
	cmd.Dir = opts.Dir
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
		out.Stdout, err = stdoutFunc(stdoutPipe, Stdout, opts.Trusted)
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
		out.Stderr, err = stderrFunc(stderrPipe, Stderr, opts.Trusted)
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
