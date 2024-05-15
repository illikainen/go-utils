package sandbox

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/illikainen/go-utils/src/iofs"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/process"
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

const activeEnv = "GO_UTILS_SANDBOX_ACTIVE"
const debugEnv = "GO_UTILS_SANDBOX_DEBUG"

func Run(opts Options) error {
	bin, err := exec.LookPath(os.Args[0])
	if err != nil || bin == os.Args[0] {
		bin, err = filepath.Abs(os.Args[0])
		if err != nil {
			return err
		}
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
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=1", activeEnv))

	if IsDebugging() {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=1", debugEnv))
	}

	log.Tracef("sandbox: execute: %s", strings.Join(args, " "))
	return cmd.Run()
}

func IsSandboxed() bool {
	return os.Getenv(activeEnv) == "1"
}

// To sandbox ourselves, we re-execute argv in a sandboxed subprocess.
// Unfortunately, Delve doesn't seem to support following subprocesses, so the
// terrible workaround used here is to check if the parent process was started
// by Delve.  If that's the case, an environment variable is introduced to the
// sandboxed subprocess to indicate that the subprocess should wait for a
// debugger to attach.
//
// FIXME: This is really ugly, there must be a better solution!
func IsDebugging() bool {
	env := os.Getenv(debugEnv)
	if env == "1" {
		return true
	}

	ppid := os.Getppid()
	if ppid < math.MinInt32 || ppid > math.MaxInt32 {
		return false
	}

	proc, err := process.NewProcess(int32(ppid))
	if err != nil {
		return false
	}

	cmdline, err := proc.CmdlineSlice()
	if err != nil {
		return false
	}

	return len(cmdline) > 0 && filepath.Base(cmdline[0]) == "dlv"
}

func AwaitDebugger() {
	log.Info("waiting for debugger to change `attached`...")

	attached := false
	for !attached {
		time.Sleep(1 * time.Second)
	}
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
