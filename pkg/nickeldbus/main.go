// Package nickeldbus implements all NickelDbus interactions of KoboMail
package nickeldbus

import (
	"os/exec"
	"strings"

	"github.com/clisboa/kobomail/pkg/helpers"
)

const binQndb = "/usr/bin/qndb"

// DesiredVersion is the NickelDbus version KoboMail works against
const DesiredVersion = "0.2.0"

// UseNickelDbus indicates if NickelDbus is in use
var UseNickelDbus = false

// IsInstalled determines if NickelDbus is installed
func IsInstalled() (installed bool) {
	const nickelDbusPath = "/mnt/onboard/.adds/nickeldbus"
	return helpers.FileExists(nickelDbusPath)
}

// GetVersion returns the current NickelDbus version
func GetVersion() (version string, err error) {
	arg1 := "-m"
	arg2 := "ndbVersion"
	cmd := exec.Command(binQndb, arg1, arg2)
	stdout, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(stdout)), nil
}

// LibraryRescanFull sends a request to rescan the library with a timeout
func LibraryRescanFull(timeout string) (stdout string, err error) {
	arg1 := "-t"
	arg2 := timeout
	arg3 := "-s"
	arg4 := "pfmDoneProcessing"
	arg5 := "-m"
	arg6 := "pfmRescanBooksFull"

	cmd := exec.Command(binQndb, arg1, arg2, arg3, arg4, arg5, arg6)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
