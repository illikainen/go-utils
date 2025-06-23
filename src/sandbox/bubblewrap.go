package sandbox

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/illikainen/go-utils/src/iofs"
	"github.com/illikainen/go-utils/src/process"
	"github.com/illikainen/go-utils/src/seq"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type BubblewrapOptions struct {
	Command          []string
	Env              []string
	ReadOnlyPaths    []string
	ReadWritePaths   []string
	DevPaths         []string
	AllowCommonPaths bool
	Tmpfs            bool
	Devtmpfs         bool
	Procfs           bool
	ShareNet         bool
	Stdin            io.Reader
	Stdout           process.OutputFunc
	Stderr           process.OutputFunc
}

type Bubblewrap struct {
	*BubblewrapOptions
	readOnlyPaths  []string
	readWritePaths []string
	devPaths       []string
}

func NewBubblewrap(opts *BubblewrapOptions) (*Bubblewrap, error) {
	b := &Bubblewrap{BubblewrapOptions: opts}

	err := b.AddReadWritePath(opts.ReadWritePaths...)
	if err != nil {
		return nil, err
	}

	err = b.AddReadOnlyPath(opts.ReadOnlyPaths...)
	if err != nil {
		return nil, err
	}

	err = b.AddDevPath(opts.DevPaths...)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bubblewrap) AddReadOnlyPath(path ...string) error {
	for _, cur := range path {
		if cur != "" {
			p, err := expand(cur)
			if err != nil {
				return errors.Wrapf(err, "bubblewrap: %s", cur)
			}

			exists, err := iofs.Exists(p)
			if err != nil {
				return err
			}
			if exists {
				b.readOnlyPaths = append(b.readOnlyPaths, p)
			}
		}
	}
	return nil
}

func (b *Bubblewrap) AddReadWritePath(path ...string) error {
	for _, cur := range path {
		if cur != "" {
			p, err := expand(cur)
			if err != nil {
				return errors.Wrapf(err, "bubblewrap: %s", cur)
			}

			exists, err := iofs.Exists(p)
			if err != nil {
				return err
			}

			for !exists {
				p = filepath.Dir(p)
				exists, err = iofs.Exists(p)
				if err != nil {
					return err
				}
			}

			b.readWritePaths = append(b.readWritePaths, p)
		}
	}
	return nil
}

func (b *Bubblewrap) AddDevPath(path ...string) error {
	for _, cur := range path {
		if cur != "" {
			p, err := expand(cur)
			if err != nil {
				return errors.Wrapf(err, "bubblewrap: %s", cur)
			}

			exists, err := iofs.Exists(p)
			if err != nil {
				return err
			}
			if exists {
				b.devPaths = append(b.devPaths, p)
			}
		}
	}
	return nil
}
func (b *Bubblewrap) SetShareNet(value bool) {
	b.ShareNet = value
}

func (b *Bubblewrap) SetStdin(r io.Reader) {
	b.Stdin = r
}

func (b *Bubblewrap) SetStdout(w process.OutputFunc) {
	b.Stdout = w
}

func (b *Bubblewrap) SetStderr(w process.OutputFunc) {
	b.Stderr = w
}

func (b *Bubblewrap) Confine() error {
	if IsSandboxed() {
		return nil
	}

	bin, err := os.Executable()
	if err != nil {
		return err
	}

	bin, err = filepath.Abs(bin)
	if err != nil {
		return err
	}

	// The current work directory needs to exist in the sandbox to support
	// relative paths like ../../foobar.  However, the real CWD isn't
	// mounted into the sandbox; instead, a tmpfs with the same path is
	// created in the sandbox.
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	args := []string{
		"bwrap",
		"--new-session",
		"--die-with-parent",
		"--unshare-user",
		"--unshare-ipc",
		"--unshare-pid",
		"--unshare-uts",
		"--unshare-cgroup",
		"--cap-drop", "ALL",
		"--tmpfs", cwd,
	}

	if b.Devtmpfs {
		args = append(args, "--dev", "/dev")
		log.Debug("bubblewrap: devtmpfs enabled")
	}

	if b.Procfs {
		args = append(args, "--proc", "/proc")
		log.Debugf("bubblewrap: procfs enabled")
	}

	if b.Tmpfs {
		args = append(args, "--tmpfs", "/tmp")
		log.Debug("bubblewrap: tmpfs enabled")
	}

	if !b.ShareNet {
		args = append(args, "--unshare-net")
	} else {
		log.Debug("bubblewrap: net enabled")
	}

	if b.AllowCommonPaths {
		args = append(
			args,
			"--ro-bind-try", "/etc/passwd", "/etc/passwd",
			"--ro-bind-try", "/etc/hosts", "/etc/hosts",
			"--ro-bind-try", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind-try", "/etc/nsswitch.conf", "/etc/nsswitch.conf",
			"--ro-bind-try", "/etc/os-release", "/etc/os-release",
			"--ro-bind-try", "/bin", "/bin",
			"--ro-bind-try", "/usr", "/usr",
			"--ro-bind-try", "/lib", "/lib",
			"--ro-bind-try", "/lib32", "/lib32",
			"--ro-bind-try", "/lib64", "/lib64",
			"--ro-bind", bin, bin,
		)
	}

	paths := []string{}
	for _, path := range b.readWritePaths {
		if !seq.Contains(paths, path) {
			args = append(args, "--bind", path, path)
			log.Debugf("bubblewrap: rw: %s", path)
			paths = append(paths, path)
		}
	}

	for _, path := range b.readOnlyPaths {
		if !seq.Contains(paths, path) {
			args = append(args, "--ro-bind", path, path)
			log.Debugf("bubblewrap: ro: %s", path)
			paths = append(paths, path)
		}
	}

	for _, path := range b.devPaths {
		if !seq.Contains(paths, path) {
			args = append(args, "--dev-bind", path, path)
			log.Debugf("bubblewrap: dev: %s", path)
			paths = append(paths, path)
		}
	}

	if len(b.Command) == 0 {
		args = append(args, bin)
		args = append(args, os.Args[1:]...)
	} else {
		args = append(args, b.Command...)
	}

	env := b.Env
	if env == nil {
		env = os.Environ()
	}

	log.Trace("bubblewrap: starting subprocess...")
	_, err = process.Exec(&process.ExecOptions{
		Command: args,
		Env:     append(env, fmt.Sprintf("%s=1", activeEnv)),
		Stdin:   b.Stdin,
		Stdout:  b.Stdout,
		Stderr:  b.Stderr,
	})
	if err != nil {
		return err
	}

	os.Exit(0) // revive:disable-line
	return nil
}
