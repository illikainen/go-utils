package flag

import (
	"fmt"

	"github.com/illikainen/go-utils/src/iofs"
	"github.com/illikainen/go-utils/src/sandbox"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
)

const (
	MustExist = 1 << iota
	MustNotExist
)

type Path struct {
	Value    string
	State    int
	Suffixes []string
	dst      *string
}

func PathVarP(flags *pflag.FlagSet, p *string, name string, shorthand string, value Path, usage string) {
	value.dst = p
	if value.Value != "" {
		lo.Must0(value.Set(value.Value))
	}

	flags.VarP(&value, name, shorthand, usage)
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
				return errors.Errorf("%s does not exist", path)
			}

			if p.State&MustNotExist == MustNotExist && exists {
				return errors.Errorf("%s must not exist", path)
			}
		}
	}

	if p.dst != nil {
		*p.dst = value
	}
	p.Value = value
	return nil
}

func (p *Path) String() string {
	return p.Value
}

func (p *Path) Type() string {
	return "path"
}
