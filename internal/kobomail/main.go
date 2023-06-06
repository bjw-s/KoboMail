// Package kobomail implements all KoboMail functionality
package kobomail

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bjw-s/kobomail/internal/config"
	"github.com/bjw-s/kobomail/pkg/imap"
	"github.com/bjw-s/kobomail/pkg/nickeldbus"
	"github.com/bjw-s/kobomail/pkg/nickelmenu"
	"github.com/bjw-s/kobomail/pkg/nickelseries"
	"github.com/bjw-s/kobomail/pkg/udev"
	"go.uber.org/zap"
)

// KoboMailConfig contains the KoboMailConfig instance used by KoboMail
var KoboMailConfig *config.Config

var useNickelDbus bool

// nickelUSBplugAddRemove simulates pugging in a USB cable
// we'll use this in case NickelDbus is not installed
func nickelUSBplugAction(action string) {
	const nickelHWstatusPipe = "/tmp/nickel-hardware-status"

	nickelPipe, _ := os.OpenFile(nickelHWstatusPipe, os.O_RDWR, os.ModeNamedPipe)
	nickelPipe.WriteString("usb plug " + action)
	nickelPipe.Close()
}

// nickelUSBplugAddRemove simulates pugging in a USB cable
// we'll use this in case NickelDbus is not installed
func nickelUSBplugAddRemove() {
	logger := zap.S()
	logger.Infow("Simulating plugging USB cable and wait 10s for the user to click on the connect button")
	nickelUSBplugAction("add")
	time.Sleep(10 * time.Second)
	logger.Infow("Simulating unplugging USB cable")
	nickelUSBplugAction("remove")
}

func prepareNickelDbusIntegration() error {
	logger := zap.S()
	// Check if NickelDbus is installed, if so then for interacting with Nickel
	// for library rescan and user notification will be handled with that
	// if not then let's use the bruteforce method of simulating the usb cable connect
	if nickeldbus.IsInstalled() {
		currentNickelDbusVersion, err := nickeldbus.GetVersion()
		if err != nil {
			return fmt.Errorf("nickelDbus version check error: %w", err)
		}

		if currentNickelDbusVersion != nickeldbus.DesiredVersion {
			logger.Warn(
				"NickelDbus version mismatch",
				zap.String("wanted", nickeldbus.DesiredVersion),
				zap.String("actual", currentNickelDbusVersion),
			)
		}
		useNickelDbus = true
	} else {
		logger.Debugw("Did not find NickelDbus")
		useNickelDbus = false
	}
	return nil
}

func prepareNickelMenuIntegration() error {
	logger := zap.S()
	if nickelmenu.ConfigFileFound() {
		if !nickelmenu.IsInstalled() || !KoboMailConfig.ApplicationConfig.CreateNickelMenuEntry {
			logger.Debugw("Unwanted NickelMenu configuration file found, removing")
			_, err := nickelmenu.DeleteConfigFile()
			if err != nil {
				return err
			}
		}
	} else {
		if nickelmenu.IsInstalled() && KoboMailConfig.ApplicationConfig.CreateNickelMenuEntry {
			logger.Debugw("NickelMenu configuration file not found, deploying template")
			_, err := nickelmenu.DeployConfigFile()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func prepareWifiTrigger() error {
	logger := zap.S()
	if KoboMailConfig.ApplicationConfig.RunOnWifiConnect {
		if !udev.RulesFileFound() {
			logger.Debugw("KoboMail uDev rules not found, deploying template")
			_, err := udev.DeployRulesFile()
			if err != nil {
				return err
			}
		}
	} else {
		if !nickelmenu.IsInstalled() {
			return fmt.Errorf("NickelMenu was not found, cannot run with manual execution type")
		}

		if udev.RulesFileFound() {
			logger.Debugw("KoboMail uDev rules found, removing")
			_, err := udev.DeleteUdevRulesFile()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// PreparePrerequisites sets up any configuration for the KoboMail integrations
func PreparePrerequisites() error {
	if err := prepareNickelDbusIntegration(); err != nil {
		return err
	}

	if err := prepareWifiTrigger(); err != nil {
		return err
	}

	if err := prepareNickelMenuIntegration(); err != nil {
		return err
	}
	return nil
}

func showDialog(message string, confirmationButton bool) {
	logger := zap.S()
	if useNickelDbus && KoboMailConfig.ApplicationConfig.ShowNotifications {
		logger.Debugw("Showing dialog", zap.String("message", message), zap.Bool("confirmation_button", confirmationButton))
		nickeldbus.DialogCreate(message)
		if confirmationButton {
			nickeldbus.DialogAddOKButton()
		}
	}
}

func updateDialog(message string, confirmationButton bool) {
	logger := zap.S()
	if useNickelDbus && KoboMailConfig.ApplicationConfig.ShowNotifications {
		logger.Debugw("Updating dialog", zap.String("message", message), zap.Bool("confirmation_button", confirmationButton))
		nickeldbus.DialogUpdate(message)
		if confirmationButton {
			nickeldbus.DialogAddOKButton()
		}
	}
}

// Run executes the main KoboMail logic
func Run() {
	logger := zap.S()

	// Show the user we are running opening a dialog
	showDialog("Starting up, please wait.", false)

	imapConnection, err := imap.ConnectToServer(KoboMailConfig.IMAPConfig.IMAPHost, KoboMailConfig.IMAPConfig.IMAPPort)
	if err != nil {
		var errMsg = fmt.Sprintf(
			"Failed to connect to %s:%v, please check internet connection",
			KoboMailConfig.IMAPConfig.IMAPHost,
			KoboMailConfig.IMAPConfig.IMAPPort,
		)
		showDialog(errMsg, true)
		logger.Fatalw(errMsg, zap.Error(err))
	}
	logger.Infow(
		"Connected to IMAP server",
		zap.String("host", KoboMailConfig.IMAPConfig.IMAPHost),
		zap.Int("port", KoboMailConfig.IMAPConfig.IMAPPort),
	)

	// Connected to the imap server, login
	if err := imapConnection.Login(KoboMailConfig.IMAPConfig.IMAPUser, string(KoboMailConfig.IMAPConfig.IMAPPwd)); err != nil {
		const errMsg = "Failed to authenticate to IMAP server"
		showDialog(errMsg+": "+err.Error(), true)
		logger.Fatalw(errMsg, zap.Error(err))
	}
	logger.Infow("Authenticated to IMAP server", zap.String("user", KoboMailConfig.IMAPConfig.IMAPUser))
	defer imapConnection.Logout()

	// Select mailbox so we can search on it
	mbox, err := imapConnection.SelectMailbox(KoboMailConfig.IMAPConfig.IMAPFolder)
	if err != nil {
		const errMsg = "Failed to select IMAP mailbox"
		showDialog(errMsg+" "+KoboMailConfig.IMAPConfig.IMAPFolder+": "+err.Error(), true)
		logger.Fatalw(errMsg, zap.Error(err))
	}
	logger.Infow("IMAP mailbox selected", zap.String("name", mbox.Name))

	// Apply the search criteria and check if there's any emails with that criteria
	if KoboMailConfig.IMAPConfig.EmailUnseen {
		imapConnection.SearchCriteria.WithoutFlags = []string{"\\Seen"}
	}
	if KoboMailConfig.IMAPConfig.EmailFlagType == config.EmailFlagTypePlus {
		criterium := strings.Replace(KoboMailConfig.IMAPConfig.IMAPUser, "@", "+"+KoboMailConfig.IMAPConfig.EmailFlag+"@", 1)
		imapConnection.SearchCriteria.Header.Add("TO", criterium)
	} else if KoboMailConfig.IMAPConfig.EmailFlagType == config.EmailFlagTypeSubject {
		criterium := KoboMailConfig.IMAPConfig.EmailFlag
		imapConnection.SearchCriteria.Header.Add("SUBJECT", criterium)
	}

	messages, err := imapConnection.CollectMessages()
	if err != nil {
		const errMsg = "Failed to fetch messages"
		showDialog(errMsg+": "+err.Error(), true)
		logger.Fatalw(errMsg, zap.Error(err))
	}
	numberOfEmailsFound := len(messages)
	logger.Infow("Fetched emails", zap.Int("number_of_emails_found", numberOfEmailsFound))

	// Process the collected messages
	if numberOfEmailsFound > 0 {
		updateDialog("Found "+strconv.Itoa(numberOfEmailsFound)+" emails to process. Please wait...", false)
	} else {
		const msg = "No emails found, nothing to be done."
		showDialog(msg, true)
		os.Exit(0)
	}

	numberOfEbooksProcessed := 0

	for _, msg := range messages {
		err := msg.FetchDetails()
		if err != nil {
			const errMsg = "Failed to process message"
			showDialog(errMsg+": "+err.Error(), true)
			logger.Fatalw(errMsg, zap.Any("message", msg), zap.Error(err))
		}
		logger.Infow("Processing message", zap.Any("message", msg))

		downloadedAttachmentsCount, err := msg.ProcessAttachments(KoboMailConfig.ProcessingConfig.Filetypes, KoboMailConfig.ApplicationConfig.LibraryPath)
		if err != nil {
			const errMsg = "Failed to process attachment"
			showDialog(errMsg+": "+err.Error(), true)
			logger.Fatalw(errMsg, zap.Error(err))
		}

		numberOfEbooksProcessed = numberOfEbooksProcessed + downloadedAttachmentsCount

		if KoboMailConfig.ProcessingConfig.EmailDelete {
			logger.Infow("Deleting message", zap.Any("message", msg))
			err := imapConnection.DeleteMessage(msg)
			if err != nil {
				const errMsg = "Failed to delete message"
				showDialog(errMsg+": "+err.Error(), true)
				logger.Fatalw(errMsg, zap.Any("message", msg), zap.Error(err))
			}
		}
	}

	if numberOfEbooksProcessed > 0 {
		if useNickelDbus {
			// Rescan the library for the new ebooks
			err := nickeldbus.LibraryRescan(30000, KoboMailConfig.ProcessingConfig.FullRescan)
			if err != nil {
				logger.Errorw("Could not update library", zap.Error(err))
			}
			logger.Debugw("Updated library")

			var msg = "Processed " + strconv.Itoa(numberOfEbooksProcessed) + " new ebooks."
			showDialog(msg, true)
			logger.Infow(msg)
		} else {
			// After finishing loading all messages simulate the USB cable connect
			// but only if there were any messages processed, no need to bug the user if there was nothing new
			nickelUSBplugAddRemove()
		}
	} else {
		const msg = "No emails found, nothing to be done."
		showDialog(msg, true)
		logger.Infow(msg)
	}
}

// Uninstall removes all KoboMail resources except the library folder
func Uninstall() {
	// Remove integrations
	udev.DeleteUdevRulesFile()
	nickelmenu.DeleteConfigFile()
	nickelseries.Uninstall()
}
