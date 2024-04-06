package cobrax

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Run(fn func(*cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		err := fn(cmd, args)
		if err != nil {
			log.Tracef("%+v", err)
			log.Fatalf("%s", err)
		}
	}
}

func ValidateArgsLength(min int, max int) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		if (min >= 0 && len(args) < min) || (max >= 0 && len(args) > max) {
			return errors.Errorf("required argument(s) not provided")
		}
		return nil
	}
}
