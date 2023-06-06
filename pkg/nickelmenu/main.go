// Package nickelmenu implements all NickelMenu interactions of KoboMail
package nickelmenu

import (
	"os"

	"github.com/bjw-s/kobomail/pkg/helpers"
	"go.uber.org/zap"
)

const nickelMenuPath = "/mnt/onboard/.adds/nm"
const nickelMenuConfigPath = nickelMenuPath + "/kobomail"

const configTemplate = "menu_item:main:KoboMail:cmd_spawn:quiet:exec usr/local/kobomail/kobomail_launcher.sh manual"

// IsInstalled determines if NickelMenu is installed
func IsInstalled() (installed bool) {
	return helpers.FolderExists(nickelMenuPath)
}

// ConfigFileFound determines if a NickelMenu configuration file is present
func ConfigFileFound() (installed bool) {
	logger := zap.S()
	configPresent := helpers.FileExists(nickelMenuConfigPath)
	logger.Debugw(
		"Checking if NickelMenu configuration file is present",
		zap.String("file", nickelMenuConfigPath),
		zap.Bool("found", configPresent),
	)
	return configPresent
}

// DeployConfigFile deploys the NickelMenu config file to the correct place so we can run KoboMail mannually
func DeployConfigFile() (ok bool, err error) {
	logger := zap.S()
	logger.Debugw(
		"Writing NickelMenu configuration file",
		zap.String("file", nickelMenuConfigPath),
	)
	err = os.WriteFile(nickelMenuConfigPath, []byte(configTemplate+"\n"), 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteConfigFile delete the NickelMenu config file if present
func DeleteConfigFile() (ok bool, err error) {
	logger := zap.S()
	logger.Debugw(
		"Removing NickelMenu configuration file",
		zap.String("file", nickelMenuConfigPath),
	)
	if ConfigFileFound() {
		_, err = helpers.DeleteFile(nickelMenuConfigPath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return true, nil
}
