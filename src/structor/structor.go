package structor

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type Options struct {
}

func Apply(value any, opts *Options) error {
	validatorSet, err := NewValidatorSet(opts)
	if err != nil {
		return err
	}

	transformerSet, err := NewTransformerSet(opts)
	if err != nil {
		return err
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		err := apply(v.Elem(), validatorSet, transformerSet, opts)
		if err != nil {
			return err
		}
	} else if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice &&
		v.Elem().Type().Elem().Kind() == reflect.Struct {
		for i := 0; i < v.Elem().Len(); i++ {
			err := apply(v.Elem().Index(i), validatorSet, transformerSet, opts)
			if err != nil {
				return err
			}
		}
	} else {
		return errors.Errorf("%v must be a pointer to a struct or a slice of structs", value)
	}

	return nil
}

func apply(v reflect.Value, validatorSet *ValidatorSet, transformerSet *TransformerSet, opts *Options) error {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldTyp := t.Field(i)
		fieldVal := v.Field(i)
		kind := fieldVal.Kind()

		if !fieldTyp.IsExported() {
			continue
		}

		_, err := setDefault(fieldTyp, fieldVal)
		if err != nil {
			return err
		}

		if kind == reflect.Struct {
			err := apply(fieldVal, validatorSet, transformerSet, opts)
			if err != nil {
				return err
			}
		} else if kind == reflect.Ptr && fieldVal.Elem().Kind() == reflect.Struct {
			err := apply(fieldVal.Elem(), validatorSet, transformerSet, opts)
			if err != nil {
				return err
			}
		} else if kind == reflect.Slice {
			for j := 0; j < fieldVal.Len(); j++ {
				v := fieldVal.Index(j)
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}

				if v.Kind() == reflect.Struct {
					err := apply(v, validatorSet, transformerSet, opts)
					if err != nil {
						return err
					}
				}
			}
		}

		validators, err := ParseValidatorTag(fieldTyp.Tag.Get("validate"))
		if err != nil {
			return err
		}

		for _, validator := range validators {
			err := validatorSet.Validate(validator.Name, fieldTyp.Name, fieldVal, validator.Args, v)
			if err != nil {
				return err
			}
		}

		transformers := fieldTyp.Tag.Get("transform")
		if transformers != "" {
			for _, transformer := range strings.Split(transformers, ",") {
				err := transformerSet.Transform(transformer, fieldTyp.Name, fieldVal)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
