package sandbox

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/illikainen/go-utils/src/iofs"
	"github.com/illikainen/go-utils/src/logging"
	"github.com/illikainen/go-utils/src/process"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	Command []string
	Env     []string
	RO      []string
	RW      []string
	Dev     []string
	Share   int
	Stdin   io.Reader
	Stdout  process.OutputFunc
	Stderr  process.OutputFunc
}

const (
	ShareNet = 1 << iota
)

const disableEnv = "GO_SANDBOX_DISABLE"
const activeEnv = "GO_SANDBOX_ACTIVE"
const debugEnv = "GO_SANDBOX_DEBUG"

func init() {
	if Compatible() && IsSandboxed() {
		// The parent process should sanitize the output by removing eg., ANSI
		// escape sequences before priting.  However, that removes styling with
		// colors etc., making it preferable to log as JSON in the sandboxed
		// subprocess and let the parent process apply the desired styling.
		log.SetFormatter(&logging.SanitizedJSONFormatter{})

		if os.Getenv(debugEnv) == "1" {
			AwaitDebugger()
		}
	}
}

func Exec(opts Options) (*process.ExecOutput, error) {
	bin, err := exec.LookPath(os.Args[0])
	if err != nil || bin == os.Args[0] {
		bin, err = filepath.Abs(os.Args[0])
		if err != nil {
			return nil, err
		}
	}

	// The current work directory needs to exist in the sandbox to support
	// relative paths like ../../foobar.  However, the real CWD isn't
	// mounted into the sandbox; instead, a tmpfs with the same path is
	// created in the sandbox.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
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
		"--dev", "/dev",
		"--tmpfs", "/tmp",
		"--tmpfs", cwd,
		"--ro-bind-try", "/etc/passwd", "/etc/passwd",
		"--ro-bind-try", "/etc/hosts", "/etc/hosts",
		"--ro-bind-try", "/etc/resolv.conf", "/etc/resolv.conf",
		"--ro-bind-try", "/etc/nsswitch.conf", "/etc/nsswitch.conf",
		"--ro-bind-try", "/usr", "/usr",
		"--ro-bind-try", "/lib", "/lib",
		"--ro-bind-try", "/lib32", "/lib32",
		"--ro-bind-try", "/lib64", "/lib64",
		"--ro-bind", bin, bin,
	}

	if opts.Share&ShareNet != ShareNet {
		args = append(args, "--unshare-net")
	}

	for _, path := range opts.RW {
		if path != "" {
			path, err := expand(path)
			if err != nil {
				return nil, err
			}
			args = append(args, "--bind-try", path, path)
		}
	}

	for _, path := range opts.RO {
		if path != "" {
			path, err = expand(path)
			if err != nil {
				return nil, err
			}
			args = append(args, "--ro-bind-try", path, path)
		}
	}

	for _, path := range opts.Dev {
		if path != "" {
			path, err = expand(path)
			if err != nil {
				return nil, err
			}

			exists, err := iofs.Exists(path)
			if err != nil {
				return nil, err
			}

			if exists {
				args = append(args, "--dev-bind", path, path)
			}
		}
	}

	args = append(args, opts.Command...)
	env := opts.Env
	if env == nil {
		env = os.Environ()
	}

	log.Trace("sandbox: starting subprocess...")
	return process.Exec(&process.ExecOptions{
		Command: args,
		Env:     append(env, fmt.Sprintf("%s=1", activeEnv)),
		Stdin:   opts.Stdin,
		Stdout:  opts.Stdout,
		Stderr:  opts.Stderr,
	})
}

func IsSandboxed() bool {
	return os.Getenv(activeEnv) == "1"
}

func AwaitDebugger() {
	log.Info("waiting for debugger to change `attached`...")

	attached := false
	for !attached {
		time.Sleep(1 * time.Second)
	}
}

func Compatible() bool {
	return os.Getenv(disableEnv) != "1" && runtime.GOOS == "linux"
}

func expand(path string) (string, error) {
	path, err := iofs.Expand(path)
	if err != nil {
		return "", err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	home = strings.TrimRight(home, string(os.PathSeparator))

	if path == home || path == string(os.PathSeparator) {
		return "", errors.Errorf("sharing %s or %s is not allowed", home, string(os.PathSeparator))
	}

	return path, nil
}
