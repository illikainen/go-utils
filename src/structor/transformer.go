package structor

import (
	"reflect"

	"github.com/pkg/errors"
)

type Transformer interface {
	Init(opts *Options) error
	Transform(name string, value reflect.Value) error
}

type TransformerSet struct {
	transformers map[string]Transformer
}

func NewTransformerSet(opts *Options) (*TransformerSet, error) {
	transformers := map[string]Transformer{
		"path": &transformPath{},
	}

	for _, transformer := range transformers {
		err := transformer.Init(opts)
		if err != nil {
			return nil, err
		}
	}

	return &TransformerSet{transformers: transformers}, nil
}

func (v *TransformerSet) Transform(name string, fieldName string, value reflect.Value) error {
	transformer, ok := v.transformers[name]
	if !ok {
		return errors.Errorf("invalid transformer: %s", name)
	}

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			v := value.Index(i)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			err := transformer.Transform(fieldName, v)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return transformer.Transform(fieldName, value)
}
