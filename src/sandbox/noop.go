package sandbox

import (
	"io"

	"github.com/illikainen/go-utils/src/process"
)

type Noop struct {
}

func NewNoop() (*Noop, error) {
	return &Noop{}, nil
}

func (n *Noop) AddReadOnlyPath(...string) error {
	return nil
}

func (n *Noop) AddReadWritePath(...string) error {
	return nil
}

func (n *Noop) AddDevPath(...string) error {
	return nil
}

func (n *Noop) SetShareNet(bool) {
}

func (n *Noop) SetStdin(io.Reader) {
}

func (n *Noop) SetStdout(process.OutputFunc) {
}

func (n *Noop) SetStderr(process.OutputFunc) {
}

func (n *Noop) Confine() error {
	return nil
}
