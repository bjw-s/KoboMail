// Package nickelseries implements all NickelSeries interactions of KoboMail
package nickelseries

import (
	"github.com/bjw-s/kobomail/pkg/helpers"
	"go.uber.org/zap"
)

const nickelSeriesPath = "/usr/local/Kobo/imageformats/libns.so"

// IsInstalled determines if NickelSeries is installed
func IsInstalled() (installed bool) {
	return helpers.FileExists(nickelSeriesPath)
}

// Uninstall delete the NickelSeries binary if present
func Uninstall() (ok bool, err error) {
	logger := zap.S()
	logger.Debugw(
		"Removing NickelSeries binary file",
		zap.String("file", nickelSeriesPath),
	)
	if IsInstalled() {
		_, err = helpers.DeleteFile(nickelSeriesPath)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return true, nil
}
