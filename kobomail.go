// The main entrypoint for KoboMail
package main

import (
	"github.com/bjw-s/kobomail/internal/config"
	"github.com/bjw-s/kobomail/internal/kobomail"
	"github.com/bjw-s/kobomail/internal/logging"
	"github.com/bjw-s/kobomail/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	var koboMailConfig = config.New()
	kobomail.KoboMailConfig = koboMailConfig

	logging.Init(
		kobomail.KoboMailConfig.ApplicationConfig.ConfigPath+"/kobomail.log",
		logging.LoglevelToZapLevel(kobomail.KoboMailConfig.ApplicationConfig.LogLevel),
	)

	logger.Debug("Running with configuration",
		zap.Any("configuration", kobomail.KoboMailConfig),
	)

	// Prepare prerequisites
	logger.Debug("Preparing prerequisites")
	if err := kobomail.PreparePrerequisites(); err != nil {
		logger.Fatal(
			"Could not prepare prerequisites",
			zap.Error(err),
		)
	}

	// Run the KoboMail processing flow
	kobomail.Run()
}
