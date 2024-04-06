package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/illikainen/go-utils/src/iofs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	Args  []string
	RO    []string
	RW    []string
	Share int
}

const (
	ShareNet = 1 << iota
)

const env = "GO_UTILS_SANDBOXED"

func Run(opts Options) error {
	bin, err := filepath.Abs(os.Args[0])
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
		"--dev", "/dev",
		"--tmpfs", "/tmp",
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
				return err
			}
			args = append(args, "--bind-try", path, path)
		}
	}

	for _, path := range opts.RO {
		if path != "" {
			path, err = expand(path)
			if err != nil {
				return err
			}
			args = append(args, "--ro-bind-try", path, path)
		}
	}

	args = append(args, opts.Args...)

	cmd := exec.Command(args[0], args[1:]...) // #nosec G204
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=1", env))

	log.Tracef("sandbox: execute: %s", strings.Join(args, " "))
	return cmd.Run()
}

func IsSandboxed() bool {
	return os.Getenv(env) == "1"
}

func Compatible() bool {
	return runtime.GOOS == "linux"
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
