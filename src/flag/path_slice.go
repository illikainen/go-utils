package flag

import (
	"strings"
)

type PathSlice struct {
	Value    []*Path
	State    int
	Suffixes []string
}

func (ps *PathSlice) Set(value string) error {
	p := &Path{
		State:    ps.State,
		Suffixes: ps.Suffixes,
	}

	ps.Value = append(ps.Value, p)
	return p.Set(value)
}

func (ps *PathSlice) String() string {
	return "[" + strings.Join(ps.StringSlice(), ", ") + "]"
}

func (ps *PathSlice) StringSlice() []string {
	values := []string{}
	for _, v := range ps.Value {
		values = append(values, v.String())
	}
	return values
}

func (ps *PathSlice) Type() string {
	return "pathslice"
}
