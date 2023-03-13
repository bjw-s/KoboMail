// Package logging configures the global logger for KoboMail
package logging

import (
	"strings"

	"github.com/bjw-s/kobomail/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init initializes a new logger at the desired log level
func Init(logFile string, logLevel zapcore.Level) {
	if err := logger.CreateComboLogger(logFile, logLevel); err != nil {
		logger.Fatal(
			"Could not prepare prerequisites",
			zap.Error(err),
		)
	}
}

// LoglevelToZapLevel converts a log level string to a Zap log level
func LoglevelToZapLevel(loglevel string) zapcore.Level {
	switch strings.ToLower(loglevel) {
	case "debug":
		return zap.DebugLevel
	case "error":
		return zap.ErrorLevel
	case "warn":
		return zap.WarnLevel
	}
	return zap.InfoLevel
}
