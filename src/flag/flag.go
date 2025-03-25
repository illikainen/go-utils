package flag

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

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
