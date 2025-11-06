package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(level zerolog.Level) zerolog.Logger {
	zerolog.TimeFieldFormat = time.DateTime

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	logger = logger.Level(level)

	return logger
}

func NewConsole(level zerolog.Level) zerolog.Logger {
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.DateTime,
	}

	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	logger = logger.Level(level)

	return logger
}
