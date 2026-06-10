package logger

import (
	"os"
	"strings"
	"time"

	"github.com/aliamerj/wardu/shared/env"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

// Setup configures zerolog for the process and returns a service-scoped logger.
func Setup(service string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = false

	level := parseLevel(env.GetString("LOG_LEVEL", "info"))
	zerolog.SetGlobalLevel(level)

	base := zerolog.New(os.Stdout)
	if env.GetBool("LOG_PRETTY", false) || strings.EqualFold(env.GetString("APP_ENV", "production"), "development") {
		base = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}

	logger := base.With().
		Timestamp().
		Str("service", service).
		Logger()

	zlog.Logger = logger
	return logger
}

// Logger returns the currently configured process logger.
func Logger() zerolog.Logger {
	return zlog.Logger
}

func parseLevel(raw string) zerolog.Level {
	level, err := zerolog.ParseLevel(strings.ToLower(strings.TrimSpace(raw)))
	if err != nil {
		return zerolog.InfoLevel
	}

	return level
}
