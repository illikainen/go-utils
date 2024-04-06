package flag

import (
	"fmt"
	"strings"

	"github.com/illikainen/go-utils/src/iofs"

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

	*p.dst = value
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
