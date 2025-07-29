package structor

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

var ErrValidation = errors.New("validation failed")

type Validator interface {
	Init(opts *Options) error
	Validate(name string, value reflect.Value, args []string, root reflect.Value) error
	NumArgs() (int, int)
}

type ValidatorSet struct {
	validators map[string]Validator
}

func NewValidatorSet(opts *Options) (*ValidatorSet, error) {
	validators := map[string]Validator{
		"base64":    &validateBase64{},
		"either":    &validateEither{},
		"email":     &validateEmail{},
		"exists":    &validateExists{},
		"group":     &validateGroup{},
		"hex":       &validateHex{},
		"host":      &validateHost{},
		"in":        &validateIn{},
		"ip":        &validateIP{},
		"max":       &validateMax{},
		"min":       &validateMin{},
		"port":      &validatePort{},
		"printable": &validatePrintable{},
		"required":  &validateRequired{},
		"rx":        &validateRegexp{},
		"uri":       &validateURI{},
		"user":      &validateUser{},
	}

	for _, validator := range validators {
		err := validator.Init(opts)
		if err != nil {
			return nil, err
		}
	}

	return &ValidatorSet{validators: validators}, nil
}

func (v *ValidatorSet) Validate(name string, fieldName string, value reflect.Value,
	args []string, root reflect.Value) error {
	validator, ok := v.validators[name]
	if !ok {
		return errors.Errorf("invalid validator: %s", name)
	}

	min, max := validator.NumArgs()
	actual := len(args)
	if actual < min || (max >= 0 && actual > max) {
		return errors.Errorf("%s: %d arguments provided to %s but only %d-%d arguments accepted",
			fieldName, actual, name, min, max)
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

			err := validator.Validate(fieldName, v, args, root)
			if err != nil {
				return err
			}
		}
		return nil
	}

	return validator.Validate(fieldName, value, args, root)
}

type ValidatorTag struct {
	Name string
	Args []string
}

func ParseValidatorTag(tag string) ([]*ValidatorTag, error) {
	var cur strings.Builder

	bracketDepth := 0
	parenDepth := 0
	quote := rune(0)
	elts := []string{}

	for i, r := range tag {
		switch r {
		case '[':
			if quote == 0 {
				bracketDepth++
			}
		case ']':
			if quote == 0 {
				bracketDepth--
			}
		case '(':
			if quote == 0 {
				parenDepth++
			}
		case ')':
			if quote == 0 {
				parenDepth--
			}
		case '\'', '"':
			if i == 0 || tag[i-1] != '\\' {
				if quote == 0 {
					quote = r
				} else if quote == r {
					quote = 0
				}
			}
		case ',':
			if bracketDepth == 0 && parenDepth == 0 && quote == 0 {
				elts = append(elts, strings.TrimSpace(cur.String()))
				cur.Reset()
				continue
			}
		}

		_, err := cur.WriteRune(r)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if cur.Len() > 0 {
		elts = append(elts, strings.TrimSpace(cur.String()))
	}

	var tags []*ValidatorTag
	for _, elt := range elts {
		parts := strings.Split(elt, ":")
		tags = append(tags, &ValidatorTag{
			Name: parts[0],
			Args: parts[1:],
		})
	}

	return tags, nil
}

func argByName(name string, args []string) string {
	prefix := name + "="
	for _, arg := range args {
		if strings.HasPrefix(arg, prefix) {
			return arg[len(prefix):]
		}
	}
	return ""
}

func argsByName(name string, args []string) []string {
	value := argByName(name, args)
	if value != "" {
		return strings.Split(value, "|")
	}
	return nil
}
