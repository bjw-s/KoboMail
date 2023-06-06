// Package config implements all commands of KoboMail
package commands

import (
	"fmt"
	"os"

	"github.com/bjw-s/kobomail/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	conf = &config.Config{}

	rootCmd = &cobra.Command{
		Use:   "kobomail",
		Short: "KoboMail is an email attachment downloader for Kobo devices",
		Long: `KoboMail is an email attachment downloader for Kobo devices.
More information available at the Github Repo (https://github.com/bjw-s/KoboMail)`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)
	cobra.OnFinalize(finalizeLogger)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringP("config", "c", config.DefaultAddonPath+"/kobomail_cfg.toml", "config.toml file for parsing authentication information")
	rootCmd.PersistentFlags().String("log-file", config.DefaultAddonPath+"/kobomail.log", "Log file location")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().String("log-format", "console", "Log format (console, json)")
	rootCmd.PersistentFlags().String("library-path", config.DefaultLibraryPath, "KoboMail library location")
}

func initConfig() {
	var err error
	conf, err = config.LoadConfig(rootCmd.PersistentFlags())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := conf.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initLogger() {
	atom := zap.NewAtomicLevel()

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	if conf.ApplicationConfig.LogFormat == "json" {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}

	var core zapcore.Core
	if conf.ApplicationConfig.LogFile != "/dev/stdout" {
		logFile, _ := os.OpenFile(conf.ApplicationConfig.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.Lock(logFile), atom),
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), atom),
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), atom),
		)
	}
	logger := zap.New(core)

	// Create a logger with a default level first to ensure config failures are loggable.
	atom.SetLevel(zapcore.InfoLevel)
	zap.ReplaceGlobals(logger)

	lvl, err := zapcore.ParseLevel(conf.ApplicationConfig.LogLevel)
	if err != nil {
		zap.S().Errorf("Invalid log level %s, using default level: info", conf.ApplicationConfig.LogLevel)
		lvl = zapcore.InfoLevel
	}
	atom.SetLevel(lvl)

	zap.S().Debug("Logger initialized")
}

func finalizeLogger() {
	// Flushes buffered log messages
	zap.S().Sync()
}
