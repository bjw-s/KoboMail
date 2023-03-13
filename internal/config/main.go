// Package config implements all configuration aspects of KoboMail
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bjw-s/kobomail/pkg/logger"
	toml "github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
)

// Set default paths
const (
	DefaultnickelHWstatusPipe = "/tmp/nickel-hardware-status"
)

type sensitiveString string

func (s sensitiveString) String() string {
	return "[REDACTED]"
}
func (s sensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

// KoboMailConfig config struct
type KoboMailConfig struct {
	IMAPConfig        imapConfig        `koanf:"imap_config" validate:"required"`
	ProcessingConfig  processingConfig  `koanf:"processing_config" validate:"required"`
	ApplicationConfig applicationConfig `koanf:"application_config" validate:"required"`
}

type imapConfig struct {
	IMAPHost      string          `koanf:"imap_host"`
	IMAPPort      int             `koanf:"imap_port"`
	IMAPUser      string          `koanf:"imap_user"`
	IMAPPwd       sensitiveString `koanf:"imap_pwd"`
	IMAPFolder    string          `koanf:"imap_folder"`
	EmailFlagType EmailFlagType   `koanf:"email_flag_type" validate:"required,oneof=plus subject"`
	EmailFlag     string          `koanf:"email_flag"`
	EmailUnseen   bool            `koanf:"email_unseen"`
}

// EmailFlagType enum
type EmailFlagType string

// EmailFlagType enum values
const (
	EmailFlagTypePlus    EmailFlagType = "plus"
	EmailFlagTypeSubject EmailFlagType = "subject"
)

type processingConfig struct {
	EmailDelete bool     `koanf:"email_delete"`
	Filetypes   []string `koanf:"filetypes"`
	FullRescan  bool     `koanf:"full_rescan"`
	Kepubify    bool     `koanf:"kepubify"`
}

type applicationConfig struct {
	CreateNickelMenuEntry bool   `koanf:"create_nickelmenu_entry"`
	RunOnWifiConnect      bool   `koanf:"run_on_wifi_connect"`
	ShowNotifications     bool   `koanf:"show_notifications"`
	ConfigPath            string `koanf:"config_path"`
	LibraryPath           string `koanf:"library_path"`
	LogLevel              string `koanf:"loglevel" validate:"required,oneof=error warn info debug"`
}

// New instantiates a new KoboMailConfig from a configFile
func New() *KoboMailConfig {
	const defaultAddonPath = "/mnt/onboard/.adds/kobomail"

	var koboMailConfig KoboMailConfig
	var k = koanf.New(".")

	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	f.String("configpath", defaultAddonPath, "Location of the KoboMail configuration files")
	f.String("loglevel", "info", "Log level to run with")
	f.Parse(os.Args[1:])

	addonPath, _ := f.GetString("configpath")
	logLevel, _ := f.GetString("loglevel")

	// Load default values using the confmap provider.
	k.Load(confmap.Provider(map[string]interface{}{
		"application_config": map[string]interface{}{
			"create_nickelmenu_entry": true,
			"config_path":             addonPath,
			"library_path":            "/mnt/onboard/KoboMailLibrary",
			"loglevel":                logLevel,
			"show_notifications":      true,
		},
		"processing_config": map[string]interface{}{
			"email_delete": false,
			"full_rescan":  false,
		},
	}, ""), nil)

	// Load configuration from config file
	if err := k.Load(file.Provider(addonPath+"/kobomail_cfg.toml"), toml.Parser()); err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	k.Unmarshal("", &koboMailConfig)

	_, validationErrors := Validate(&koboMailConfig)
	if validationErrors != nil {
		for _, ve := range validationErrors {
			logger.Error(
				"Configuration validation failed",
				zap.String("field", ve.Namespace()),
				zap.String("error", errorMessageForValidationError(ve)),
			)
		}
		os.Exit(1)
	}

	return &koboMailConfig
}
