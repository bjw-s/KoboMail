// Package nickelmenu implements all NickelMenu interactions of KoboMail
package nickelmenu

import (
	"os"

	"github.com/clisboa/kobomail/pkg/helpers"
)

const nickelMenuPath = "/mnt/onboard/.adds/nm"
const nickelMenuConfigPath = nickelMenuPath + "/kobomail"

const configTemplate = "menu_item:main:KoboMail:cmd_spawn:quiet:exec usr/local/kobomail/kobomail_launcher.sh manual"

// IsInstalled determines if NickelMenu is installed
func IsInstalled() (installed bool) {
	return helpers.FileExists(nickelMenuPath)
}

// ConfigFileFound determines if a NickelMenu configuration file is present
func ConfigFileFound() (installed bool) {
	return helpers.FileExists(nickelMenuConfigPath)
}

// DeployConfigFile deploys the NickelMenu config file to the correct place so we can run KoboMail mannually
func DeployConfigFile() (ok bool, err error) {
	err = os.WriteFile(nickelMenuConfigPath, []byte(configTemplate+"\n"), 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteConfigFile delete the NickelMenu config file if present
func DeleteConfigFile() (ok bool, err error) {
	if ConfigFileFound() {
		err = os.Remove(nickelMenuConfigPath)

		if err != nil {
			return false, err
		}
		return true, nil
	}
	return true, nil
}
