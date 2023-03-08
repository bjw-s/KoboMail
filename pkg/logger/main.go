// Package logger implements all logging aspects of KoboMail
package logger

import (
	"os"
	"time"

	"github.com/clisboa/kobomail/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is a global instance of the Zap logger
var Log *zap.Logger

func init() {
	logFile, err := os.OpenFile(config.DefaultPath+"/kobomail.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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

	Log = zap.New(core)
	defer Log.Sync()
}

// Info logs a message at level Info on the zap logger
func Info(message string, fields ...zap.Field) {
	Log.Info(message, fields...)
}

// Debug logs a message at level Debug on the zap logger
func Debug(message string, fields ...zap.Field) {
	Log.Debug(message, fields...)
}

// Warn logs a message at level Warning on the zap logger
func Warn(message string, fields ...zap.Field) {
	Log.Warn(message, fields...)
}

// Error logs a message at level Error on the zap logger
func Error(message string, fields ...zap.Field) {
	Log.Error(message, fields...)
}

// Fatal logs a message at level Fatal on the zap logger
func Fatal(message string, fields ...zap.Field) {
	Log.Fatal(message, fields...)
}
