package logging

import (
	"fmt"

	"github.com/illikainen/go-utils/src/stringx"

	"github.com/fatih/color"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
)

func WithSuppress(fn func() error) error {
	level := log.GetLevel()
	log.SetLevel(log.FatalLevel)

	err := fn()
	log.SetLevel(level)
	return err
}

type SanitizedTextFormatter struct {
}

func (f *SanitizedTextFormatter) Format(entry *log.Entry) ([]byte, error) {
	level := ""

	switch entry.Level {
	case log.TraceLevel:
		level = color.CyanString(entry.Level.String())
	case log.DebugLevel:
		level = color.WhiteString(entry.Level.String())
	case log.InfoLevel:
		level = color.GreenString(entry.Level.String())
	case log.WarnLevel:
		level = color.YellowString(entry.Level.String())
	default:
		level = color.RedString(entry.Level.String())
	}

	return lo.FlatMap(stringx.SplitLines(entry.Message), func(line string, _ int) []byte {
		return []byte(fmt.Sprintf("%-14s | %s\n", level, stringx.Sanitize(line)))
	}), nil
}
