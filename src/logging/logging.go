package logging

import (
	"fmt"

	"github.com/illikainen/go-utils/src/seq"
	"github.com/illikainen/go-utils/src/stringx"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

func WithSuppress(fn func() error) error {
	level := log.GetLevel()
	log.SetLevel(log.FatalLevel)

	err := fn()
	log.SetLevel(level)
	return err
}

type SanitizedJSONFormatter struct {
}

func (f *SanitizedJSONFormatter) Format(e *log.Entry) ([]byte, error) {
	formatter := log.JSONFormatter{}
	out, err := formatter.Format(e)
	if err != nil {
		return nil, err
	}
	return stringx.Sanitize(out), nil
}

type SanitizedTextFormatter struct {
}

func (f *SanitizedTextFormatter) Format(entry *log.Entry) ([]byte, error) {
	if GetField(entry.Data, "unstyled", false) {
		return []byte(stringx.Sanitize(entry.Message) + "\n"), nil
	}

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

	return seq.ExpandBy(stringx.SplitLines(entry.Message), func(line string, _ int) []byte {
		return []byte(fmt.Sprintf("%-14s | %s\n", level, stringx.Sanitize(line)))
	}), nil
}

func GetField[T bool | string](fields log.Fields, key string, fallback T) T {
	value, ok := fields[key]
	if !ok {
		return fallback
	}

	v, ok := value.(T)
	if !ok {
		return fallback
	}

	return v
}
