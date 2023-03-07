// Package udev implements all udev interactions of KoboMail
package udev

import "strings"

type udevRules []string

func (r udevRules) generateFile() string {
	return strings.Join(r, "\n") + "\n"
}
