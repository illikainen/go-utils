package logging

import (
	"io"

	log "github.com/sirupsen/logrus"
)

type Logger interface {
	Trace(...any)
	Tracef(string, ...any)
	Debug(...any)
	Debugf(string, ...any)
	Info(...any)
	Infof(string, ...any)
	Warn(...any)
	Warnf(string, ...any)
	Error(...any)
	Errorf(string, ...any)
	Fatal(...any)
	Fatalf(string, ...any)
}

func DiscardLogger() Logger {
	logger := log.New()
	logger.SetFormatter(&SanitizedTextFormatter{})
	logger.SetLevel(log.FatalLevel)
	logger.SetOutput(io.Discard)
	return logger
}
