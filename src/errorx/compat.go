package errorx

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Join is copied from Go 1.20.1.
//
// """
// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// """
//
// The LICENSE file can be found in the Go codebase.
//
// TODO: replace with `errors.Join()` from stdlib when Debian 13 is released.
func Join(errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	e := &joinError{
		errs: make([]error, 0, n),
	}
	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
	return e
}

type joinError struct {
	errs []error
}

func (e *joinError) Error() string {
	strs := []string{}
	for _, err := range e.errs {
		strs = append(strs, err.Error())
	}
	return strings.Join(strs, "\n")
}

func (e *joinError) Unwrap() error {
	lines := []string{}
	for _, err := range e.errs {
		lines = append(lines, fmt.Sprintf("%s", err))
	}
	return errors.Errorf("%s", strings.Join(lines, "\n"))
}

func (e *joinError) UnwrapSlice() []error {
	return e.errs
}
