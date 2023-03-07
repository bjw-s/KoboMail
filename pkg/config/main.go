// Package config implements all configuration aspects of KoboMail
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	toml "github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
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
	IMAPConfig       imapConfig       `koanf:"imap_config" validate:"required"`
	ExecutionType    executionType    `koanf:"execution_type" validate:"required"`
	ProcessingConfig processingConfig `koanf:"processing_config"`
}

type imapConfig struct {
	IMAPHost      string          `koanf:"imap_host"`
	IMAPPort      string          `koanf:"imap_port"`
	IMAPUser      string          `koanf:"imap_user"`
	IMAPPwd       sensitiveString `koanf:"imap_pwd"`
	IMAPFolder    string          `koanf:"imap_folder"`
	EmailFlagType EmailFlagType   `koanf:"email_flag_type" validate:"required,oneof=plus subject"`
	EmailFlag     string          `koanf:"email_flag"`
	EmailUnseen   string          `koanf:"email_unseen"`
	EmailDelete   string          `koanf:"email_delete"`
}

type executionType struct {
	Type ExecutionType `koanf:"type" validate:"required,oneof=auto manual"`
}

// ExecutionType enum
type ExecutionType string

// ExecutionType enum values
const (
	ExecutionTypeManual ExecutionType = "manual"
	ExecutionTypeAuto   ExecutionType = "auto"
)

// EmailFlagType enum
type EmailFlagType string

// EmailFlagType enum values
const (
	EmailFlagTypePlus    EmailFlagType = "plus"
	EmailFlagTypeSubject EmailFlagType = "subject"
)

type processingConfig struct {
	Filetypes []string `koanf:"filetypes"`
	Kepubify  string   `koanf:"kepubify"`
}

// New instantiates a new KoboMailConfig from a configFile
func New(configFile string) *KoboMailConfig {
	var koboMailConfig KoboMailConfig
	var k = koanf.New(".")

	if err := k.Load(file.Provider(configFile), toml.Parser()); err != nil {
		log.Fatalf("Failed to load configuration: %v\n", err)
	}

	k.Unmarshal("", &koboMailConfig)

	validate := validator.New()
	err := validate.Struct(&koboMailConfig)
	if err != nil {

		for _, fe := range err.(validator.ValidationErrors) {
			log.Printf(
				"Configuration validation failed {\"field\": \"%s\", \"error\": \"%s\"}",
				fe.Namespace(),
				msgForTag(fe),
			)
		}
		os.Exit(1)
	}

	return &koboMailConfig
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email"
	case "oneof":
		return fmt.Sprintf("Invalid value. Expected one of %v", strings.Split(fe.Param(), " "))
	}
	return fe.Error() // default error
}
