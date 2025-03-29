package flag

import (
	"fmt"

	"github.com/illikainen/go-utils/src/iofs"
	"github.com/illikainen/go-utils/src/sandbox"

	"github.com/pkg/errors"
)

const (
	MustExist = 1 << iota
	MustNotExist
)

type Path struct {
	Value    string
	Values   []string
	State    int
	Suffixes []string
}

func (p *Path) Set(value string) error {
	paths := []string{}
	if len(p.Suffixes) == 0 {
		paths = append(paths, value)
	} else {
		for _, suffix := range p.Suffixes {
			paths = append(paths, fmt.Sprintf("%s.%s", value, suffix))
		}
	}

	p.Value = value
	p.Values = paths

	// The state is only validated in a non-sandboxed parent process
	// because a non-existent path has to be created before the sandboxed
	// subprocess is spawned if the file should be mounted in the sandbox.
	if !sandbox.IsSandboxed() {
		for _, path := range paths {
			exists, err := iofs.Exists(path)
			if err != nil {
				return err
			}

			if p.State&MustExist == MustExist && !exists {
				return errors.Wrap(os.ErrNotExist, path)
			}

			if p.State&MustNotExist == MustNotExist && exists {
				return errors.Wrap(os.ErrExist, path)
			}
		}
	}

	return nil
}

func (p *Path) String() string {
	return p.Value
}

func (p *Path) Type() string {
	return "path"
}
