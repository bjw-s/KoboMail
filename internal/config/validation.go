// Package config implements all configuration aspects of KoboMail
package config

import (
	"github.com/bjw-s/kobomail/pkg/helpers"
	"github.com/gookit/validate"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slices"
)

// ValidateLogLevel validates that the log level is one of the valid log levels
func (c Config) ValidateLogLevel(val string) bool {
	validLogLevels := []string{}
	for i := zapcore.DebugLevel; i < zapcore.InvalidLevel; i++ {
		validLogLevels = append(validLogLevels, i.String())
	}
	return slices.Contains(validLogLevels, val)
}

// ValidateFolder validates that the path is a valid folder
func (c Config) ValidateFolder(val string) bool {
	return helpers.FolderExists(val)
}

// Validate returns if the given configuration is valid and any validation errors
func (c *Config) Validate() validate.Errors {
	v := validate.Struct(c)
	v.StopOnError = false
	return v.ValidateE()
}

func (c Config) Messages() map[string]string {
	return validate.MS{
		"ValidateFolder": "{field} must point to a valid folder.",
		"ApplicationConfig.LogLevel.ValidateLogLevel": "Log Level must be one of: debug, info, warn, error, dpanic, panic, fatal",
	}
}
