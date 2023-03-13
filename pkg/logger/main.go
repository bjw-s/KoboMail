// Package logger implements all logging aspects of KoboMail
package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.Logger

var encoderConfig zapcore.EncoderConfig
var fileEncoder zapcore.Encoder
var consoleEncoder zapcore.Encoder
var atomicLevel zap.AtomicLevel

func init() {
	// This sets up a default zapLogger
	encoderConfig = zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	encoderConfig.StacktraceKey = ""

	fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	atomicLevel = zap.NewAtomicLevel()

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel),
	)

	zapLog = zap.New(core)
	defer zapLog.Sync()
}

// CreateComboLogger configures the logger to output to stdout and a specified file
func CreateComboLogger(logFilePath string, logLevel zapcore.Level) error {
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), atomicLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel),
	)

	zapLog = zap.New(core)
	defer zapLog.Sync()
	SetLevel(logLevel)
	return nil
}

// Info logs a message at level Info on the zap logger
func Info(message string, fields ...zap.Field) {
	zapLog.Info(message, fields...)
}

// Debug logs a message at level Debug on the zap logger
func Debug(message string, fields ...zap.Field) {
	zapLog.Debug(message, fields...)
}

// Warn logs a message at level Warning on the zap logger
func Warn(message string, fields ...zap.Field) {
	zapLog.Warn(message, fields...)
}

// Error logs a message at level Error on the zap logger
func Error(message string, fields ...zap.Field) {
	zapLog.Error(message, fields...)
}

// Fatal logs a message at level Fatal on the zap logger
func Fatal(message string, fields ...zap.Field) {
	zapLog.Fatal(message, fields...)
}

// SetLevel (re-)configures the log level
func SetLevel(logLevel zapcore.Level) {
	atomicLevel.SetLevel(logLevel)
}
