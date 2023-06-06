// Package udev implements all udev interactions of KoboMail
package udev

import (
	"os"

	"github.com/bjw-s/kobomail/pkg/helpers"
	"go.uber.org/zap"
)

var kobomailRules = udevRules{
	`KERNEL=="eth*", ACTION=="add", RUN+="/usr/local/kobomail/kobomail_launcher.sh`,
	`KERNEL=="wlan*", ACTION=="add", RUN+="/usr/local/kobomail/kobomail_launcher.sh"`,
	`KERNEL=="lo", RUN+="/usr/local/kobomail/kobomail_config_setup.sh"`,
}

const udevRulesFilePath = "/etc/udev/rules.d/97-kobomail.rules"

// RulesFileFound checks if udev rules file is present
func RulesFileFound() (installed bool) {
	logger := zap.S()
	rulesPresent := helpers.FileExists(udevRulesFilePath)
	logger.Debugw(
		"Checking if udev rules file is present",
		zap.String("file", udevRulesFilePath),
		zap.Bool("found", rulesPresent),
	)
	return rulesPresent
}

// DeployRulesFile deploys the udev rulesfile at the correct place so KoboMail runs automatically everytime WIfi is activated
func DeployRulesFile() (ok bool, err error) {
	logger := zap.S()
	logger.Debugw(
		"Writing udev rules file",
		zap.String("file", udevRulesFilePath),
	)
	err = os.WriteFile(udevRulesFilePath, []byte(kobomailRules.generateFile()), 0644)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteUdevRulesFile deletes the udev rules file if present
func DeleteUdevRulesFile() (ok bool, err error) {
	logger := zap.S()
	logger.Debugw(
		"Removing udev rules file",
		zap.String("file", udevRulesFilePath),
	)
	if RulesFileFound() {
		_, err = helpers.DeleteFile(udevRulesFilePath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return true, nil
}
