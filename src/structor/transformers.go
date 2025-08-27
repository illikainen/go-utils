package structor

import (
	"reflect"

	"github.com/illikainen/go-utils/src/iofs"

	"github.com/pkg/errors"
)

// ----
// Path
// ----
type transformPath struct {
}

func (t *transformPath) Init(_ *Options) error {
	return nil
}

func (t *transformPath) Transform(name string, value reflect.Value) error {
	if !value.IsValid() || value.IsZero() {
		return nil
	}

	if value.Kind() != reflect.String {
		return errors.Errorf("%s must be a string", name)
	}

	path, err := iofs.Expand(value.String())
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(path))
	return nil
}
