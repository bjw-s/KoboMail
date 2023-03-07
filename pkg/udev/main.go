// Package udev implements all udev interactions of KoboMail
package udev

import (
	"os"

	"github.com/clisboa/kobomail/pkg/helpers"
)

// const rulesFilePath = "/etc/udev/rules.d/97-kobomail.rules"
const rulesFilePath = "/tmp/97-kobomail.rules"

var kobomailRules = udevRules{
	`KERNEL=="eth*", ACTION=="add", RUN+="/usr/local/kobomail/kobomail_launcher.sh`,
	`KERNEL=="wlan*", ACTION=="add", RUN+="/usr/local/kobomail/kobomail_launcher.sh"`,
	`KERNEL=="lo", RUN+="/usr/local/kobomail/kobomail_config_setup.sh"`,
}

// RulesFileFound checks if udev rules file is present
func RulesFileFound() (installed bool) {
	return helpers.FileExists(rulesFilePath)
}

// DeployRulesFile deploys the udev rulesfile at the correct place so KoboMail runs automatically everytime WIfi is activated
func DeployRulesFile() (ok bool, err error) {
	err = os.WriteFile(rulesFilePath, []byte(kobomailRules.generateFile()), 0644)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteUdevRulesFile deletes the udev rules file if present
func DeleteUdevRulesFile() (ok bool, err error) {
	if RulesFileFound() {
		err = os.Remove(rulesFilePath)

		if err != nil {
			return false, err
		}
		return true, nil
	}
	return true, nil
}
