package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	*zerolog.Logger
}

func New(environment string) *Logger {
	var output io.Writer = os.Stdout

	// Pretty logging for development
	if environment == "development" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if environment == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = logger

	return &Logger{&logger}
}

func (l *Logger) WithFields(fields map[string]interface{}) *zerolog.Logger {
	ctx := l.Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	logger := ctx.Logger()
	return &logger
}

func (l *Logger) WithRequestID(requestID string) *zerolog.Logger {
	logger := l.Logger.With().Str("request_id", requestID).Logger()
	return &logger
}

func (l *Logger) WithUserID(userID string) *zerolog.Logger {
	logger := l.Logger.With().Str("user_id", userID).Logger()
	return &logger
}
