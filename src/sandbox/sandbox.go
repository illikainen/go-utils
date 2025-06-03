package sandbox

import (
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/illikainen/go-utils/src/iofs"
	"github.com/illikainen/go-utils/src/logging"
	"github.com/illikainen/go-utils/src/process"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	BubblewrapSandbox = 1 << iota
	NoSandbox
)

type Sandbox interface {
	Confine() error
	AddReadOnlyPath(path ...string) error
	AddReadWritePath(path ...string) error
	AddDevPath(path ...string) error
	SetShareNet(bool)
	SetStdin(io.Reader)
	SetStdout(process.OutputFunc)
	SetStderr(process.OutputFunc)
}

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

func Backend(name string) (int, error) {
	switch strings.ToLower(name) {
	case "bubblewrap":
		return BubblewrapSandbox, nil
	case "none":
		return NoSandbox, nil
	case "":
	default:
		return -1, errors.Errorf("%s is not a supported sandbox backend", name)
	}

	if runtime.GOOS == "linux" {
		return BubblewrapSandbox, nil
	}

	log.Warnf("sandbox not compatible with %s", runtime.GOOS)
	log.Warnf("configure `sandbox = none` to disable this warning")
	return NoSandbox, nil
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
	isDocker, err := iofs.Exists("/.dockerenv")
	if err != nil {
		return true
	}

	isPodman, err := iofs.Exists("/run/.containerenv")
	if err != nil {
		return true
	}

	return os.Getenv(disableEnv) != "1" && runtime.GOOS == "linux" && !isDocker && !isPodman
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
