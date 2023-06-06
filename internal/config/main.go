// Package config implements all configuration aspects of KoboMail
package config

import (
	"encoding/json"

	toml "github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

// Set default paths
const (
	DefaultAddonPath   = "/mnt/onboard/.adds/kobomail"
	DefaultLibraryPath = "/mnt/onboard/KoboMailLibrary"
)

type sensitiveString string

func (s sensitiveString) String() string {
	return "[REDACTED]"
}
func (s sensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

// Config config struct
type Config struct {
	IMAPConfig        imapConfigSection        `koanf:"imap_config" validate:"required"`
	ProcessingConfig  processingConfigSection  `koanf:"processing_config" validate:"required"`
	ApplicationConfig applicationConfigSection `koanf:"application_config" validate:"required"`
	k                 *koanf.Koanf
}

type imapConfigSection struct {
	IMAPHost      string          `koanf:"imap_host"`
	IMAPPort      int             `koanf:"imap_port"`
	IMAPUser      string          `koanf:"imap_user"`
	IMAPPwd       sensitiveString `koanf:"imap_pwd"`
	IMAPFolder    string          `koanf:"imap_folder"`
	EmailFlagType EmailFlagType   `koanf:"email_flag_type" validate:"required|in:plus,subject"`
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

type processingConfigSection struct {
	EmailDelete bool     `koanf:"email_delete"`
	Filetypes   []string `koanf:"filetypes"`
	FullRescan  bool     `koanf:"full_rescan"`
	Kepubify    bool     `koanf:"kepubify"`
}

type applicationConfigSection struct {
	CreateNickelMenuEntry bool   `koanf:"create_nickelmenu_entry"`
	RunOnWifiConnect      bool   `koanf:"run_on_wifi_connect"`
	ShowNotifications     bool   `koanf:"show_notifications"`
	ConfigPath            string `koanf:"config_path" validate:"ValidateFolder"`
	LibraryPath           string `koanf:"library_path" validate:"ValidateFolder"`
	LogFile               string `koanf:"logfile"`
	LogFormat             string `koanf:"logformat" validate:"in:console,json"`
	LogLevel              string `koanf:"loglevel" validate:"ValidateLogLevel"`
}

// LoadConfig instantiates a new Config
func LoadConfig(flags *flag.FlagSet) (*Config, error) {
	var err error
	var k = koanf.New(".")

	// Fetch flags
	if err = k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
		return nil, err
	}

	// Defaults
	err = k.Load(confmap.Provider(map[string]interface{}{
		"application_config": map[string]interface{}{
			"create_nickelmenu_entry": true,
			"library_path":            DefaultLibraryPath,
			"show_notifications":      true,
		},
		"processing_config": map[string]interface{}{
			"email_delete": false,
			"full_rescan":  false,
		},
	}, ""), nil)
	if err != nil {
		return nil, err
	}

	// TOML Config
	tomlConfig := k.String("config")
	if tomlConfig != "" {
		err = k.Load(file.Provider(tomlConfig), toml.Parser())
		if err != nil {
			return nil, err
		}
	}

	// Flag overrides
	err = k.Load(confmap.Provider(map[string]interface{}{
		"application_config": map[string]interface{}{
			"library_path": k.String("library-path"),
			"logfile":      k.String("log-file"),
			"loglevel":     k.String("log-level"),
			"logformat":    k.String("log-format"),
		},
	}, ""), nil)
	if err != nil {
		return nil, err
	}

	var out Config
	err = k.Unmarshal("", &out)
	if err != nil {
		return nil, err
	}

	out.k = k
	return &out, nil
}
