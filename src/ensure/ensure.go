package ensure

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// FIXME: currently only compatible with POSIX systems.
func Unprivileged() {
	if os.Getuid() == 0 || os.Geteuid() == 0 || os.Getgid() == 0 || os.Getegid() == 0 {
		log.Fatalf("%s shouldn't be launched as root", os.Args[0])
	}
}
