// Package nickeldbus implements all NickelDbus interactions of KoboMail
package nickeldbus

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"go.uber.org/zap"
)

const (
	ndbInterface  = "com.github.shermp.nickeldbus"
	ndbObjectPath = "/nickeldbus"
)

// DesiredVersion is the NickelDbus version KoboMail works against
const DesiredVersion = "0.2.0"

func getSystemDbusConnection() (*dbus.Conn, error) {
	var err error
	var ndbConn *dbus.Conn

	ndbConn, err = dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	return ndbConn, nil
}

func getNdbObject(conn *dbus.Conn) (dbus.BusObject, error) {
	var err error
	if conn == nil {
		conn, err = getSystemDbusConnection()
		if err != nil {
			return nil, err
		}
	}
	return conn.Object(ndbInterface, ndbObjectPath), nil
}

// IsInstalled returns if NickelDbus is installed
func IsInstalled() bool {
	logger := zap.S()
	var err error
	var ndbObj dbus.BusObject

	ndbObj, err = getNdbObject(nil)
	if err != nil {
		return false
	}

	_, err = introspect.Call(ndbObj)
	installed := err == nil

	logger.Debugw("NickelDbus install check", zap.Bool("installed", installed))
	return installed
}

// GetVersion returns the current NickelDbus version
func GetVersion() (string, error) {
	logger := zap.S()
	var err error
	var ndbVersion string
	var ndbObj dbus.BusObject

	ndbObj, err = getNdbObject(nil)
	if err != nil {
		return "", err
	}

	err = ndbObj.Call(ndbInterface+".ndbVersion", 0).Store(&ndbVersion)
	if err != nil {
		return "", err
	}

	logger.Debugw("NickelDbus version", zap.String("version", ndbVersion))
	return ndbVersion, nil
}
