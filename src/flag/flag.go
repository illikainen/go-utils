package flag

import (
	"fmt"
	"strings"

	"github.com/illikainen/go-utils/src/iofs"
	"github.com/illikainen/go-utils/src/sandbox"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
)

type EnumType interface {
	int | string
}

type Enum[T EnumType] struct {
	Name  string
	Value T
}

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

func SetFallback[T string | []string](flags *pflag.FlagSet, name string, value ...T) error {
	f := flags.Lookup(name)
	if f != nil && !flags.Changed(name) {
		for _, value := range value {
			switch value := any(value).(type) {
			case string:
				if value != "" {
					return f.Value.Set(value)
				}
			case []string:
				if value != nil {
					for _, v := range value {
						err := f.Value.Set(v)
						if err != nil {
							return err
						}
					}
					return nil
				}
			default:
				return errors.Errorf("invalid value type")
			}
		}
	}
	return nil
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

type StringToInt[T EnumType] struct {
	Value         string
	Kind          string
	Choices       map[string]T
	CaseSensitive bool
	dst           *Enum[T]
}

func StringToVarP[T EnumType](flags *pflag.FlagSet, p *Enum[T], name string,
	shorthand string, value StringToInt[T], usage string) {
	value.dst = p
	if value.Value != "" {
		lo.Must0(value.Set(value.Value))
	}

	flags.VarP(
		&value,
		name,
		shorthand,
		fmt.Sprintf("%s (%s)", usage, strings.Join(lo.Keys(value.Choices), ", ")),
	)
}

func (p *StringToInt[T]) Set(value string) error {
	n, ok := p.Choices[lo.Ternary(p.CaseSensitive, strings.ToUpper(value), value)]
	if !ok {
		return errors.Errorf("choices: %s", strings.Join(lo.Keys(p.Choices), ", "))
	}

	p.dst.Name = value
	p.dst.Value = n

	p.Value = value
	return nil
}

func (p *StringToInt[T]) String() string {
	return p.Value
}

func (p *StringToInt[T]) Type() string {
	return lo.Ternary(p.Kind == "", "intmap", p.Kind)
}
