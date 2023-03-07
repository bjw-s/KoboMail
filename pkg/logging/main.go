// Package logging implements all logging aspects of KoboMail
package logging

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a generic wrapper around the Zap logger
type Logger struct {
	logger *zap.Logger
}

// New instantiates a new logger
func New(file string) *Logger {
	var zapLog *zap.Logger
	logFile, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	encoderConfig.StacktraceKey = ""
	atomicLevel := zapcore.DebugLevel

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), atomicLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel),
	)

	zapLog = zap.New(core)
	defer zapLog.Sync()

	return &Logger{
		logger: zapLog,
	}
}

// Info logs a message at level Info on the zap logger
func (logger *Logger) Info(message string, fields ...zap.Field) {
	logger.logger.Info(message, fields...)
}

// Debug logs a message at level Debug on the zap logger
func (logger *Logger) Debug(message string, fields ...zap.Field) {
	logger.logger.Debug(message, fields...)
}

// Warn logs a message at level Warning on the zap logger
func (logger *Logger) Warn(message string, fields ...zap.Field) {
	logger.logger.Warn(message, fields...)
}

// Error logs a message at level Error on the zap logger
func (logger *Logger) Error(message string, fields ...zap.Field) {
	logger.logger.Error(message, fields...)
}

// Fatal logs a message at level Fatal on the zap logger
func (logger *Logger) Fatal(message string, fields ...zap.Field) {
	logger.logger.Fatal(message, fields...)
}
