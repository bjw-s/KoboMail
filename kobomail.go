// The main entrypoint for KoboMail
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/clisboa/kobomail/pkg/config"
	"github.com/clisboa/kobomail/pkg/imap"
	"github.com/clisboa/kobomail/pkg/logger"
	"github.com/clisboa/kobomail/pkg/nickeldbus"
	"github.com/clisboa/kobomail/pkg/nickelmenu"
	"github.com/clisboa/kobomail/pkg/udev"
	"go.uber.org/zap"
)

var koboMailConfig = config.New(config.DefaultPath + "/kobomail_cfg.toml")

// nickelUSBplugAddRemove simulates pugging in a USB cable
// we'll use this in case NickelDbus is not installed
func nickelUSBplugAddRemove(action string) {
	nickelPipe, _ := os.OpenFile(config.DefaultnickelHWstatusPipe, os.O_RDWR, os.ModeNamedPipe)
	nickelPipe.WriteString("usb plug " + action)
	nickelPipe.Close()
}

func preparePrerequisites(et config.ExecutionType) error {
	if et == config.ExecutionTypeAuto {
		// KoboMail relies on udev rules In AUTO mode, so make sure they are in place
		// and remove any NickelMenu configuration

		if !udev.RulesFileFound() {
			logger.Debug("KoboMail uDev rules not found, deploying template")
			_, err := udev.DeployRulesFile()
			if err != nil {
				return err
			}
		}

		if nickelmenu.IsInstalled() && nickelmenu.ConfigFileFound() {
			logger.Debug("Deleting NickelMenu configuration file")
			_, err := nickelmenu.DeleteConfigFile()
			if err != nil {
				return err
			}
		}
		return nil

	} else if et == config.ExecutionTypeManual {
		// KoboMail relies on NickelMenu In MANUAL mode, so make sure that is in place
		// and remove any udev rules

		if !nickelmenu.IsInstalled() {
			return fmt.Errorf("NickelMenu was not found, cannot run with manual execution type")
		}

		if !nickelmenu.ConfigFileFound() {
			logger.Debug("NickelMenu configuration file not found, deploying template")
			_, err := nickelmenu.DeployConfigFile()
			if err != nil {
				return err
			}
		}

		if udev.RulesFileFound() {
			logger.Debug("KoboMail uDev rules found, removing")
			_, err := udev.DeleteUdevRulesFile()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	logger.Debug("Running with configuration",
		zap.Any("configuration", koboMailConfig),
	)

	// Check if NickelDbus is installed, if so then for interacting with Nickel
	// for library rescan and user notification will be handled with that
	// if not then let's use the bruteforce method of simulating the usb cable connect
	if nickeldbus.IsInstalled() {
		currentNickelDbusVersion, err := nickeldbus.GetVersion()
		if err != nil {
			logger.Fatal("NickelDbus version check", zap.Error(err))
		}

		if currentNickelDbusVersion != nickeldbus.DesiredVersion {
			logger.Warn(
				"NickelDbus version mismatch",
				zap.String("wanted", nickeldbus.DesiredVersion),
				zap.String("actual", currentNickelDbusVersion),
			)
		} else {
			nickeldbus.UseNickelDbus = true
		}
	} else {
		logger.Info("Did not find NickelDbus")
	}

	// Show the user we are running opening a dialog
	nickeldbus.DialogCreate("Starting up, please wait.")

	// Prepare prerequisites
	logger.Debug("Preparing prerequisites", zap.String("execution_type", string(koboMailConfig.ExecutionType.Type)))
	if err := preparePrerequisites(koboMailConfig.ExecutionType.Type); err != nil {
		logger.Fatal(
			"Could not prepare prerequisites",
			zap.Error(err),
		)
	}

	imapConnection, err := imap.ConnectToServer(koboMailConfig.IMAPConfig.IMAPHost, koboMailConfig.IMAPConfig.IMAPPort)
	if err != nil {
		nickeldbus.DialogAddOKButton()
		nickeldbus.DialogUpdate(fmt.Sprintf(
			"Failed to connect to %s:%s, please check internet connection",
			koboMailConfig.IMAPConfig.IMAPHost,
			koboMailConfig.IMAPConfig.IMAPPort,
		))

		logger.Error(
			"Failed to connect to %s:%s, please check internet connection",
			zap.String("host", koboMailConfig.IMAPConfig.IMAPHost),
			zap.String("port", koboMailConfig.IMAPConfig.IMAPPort),
		)
		logger.Fatal("Error", zap.Error(err))
	}
	logger.Info(
		"Connected to IMAP server",
		zap.String("host", koboMailConfig.IMAPConfig.IMAPHost),
		zap.String("port", koboMailConfig.IMAPConfig.IMAPPort),
	)

	// Connected to the imap server, login
	if err := imapConnection.Login(koboMailConfig.IMAPConfig.IMAPUser, string(koboMailConfig.IMAPConfig.IMAPPwd)); err != nil {
		nickeldbus.DialogAddOKButton()
		nickeldbus.DialogUpdate("Failed to authenticate to IMAP server: " + err.Error())

		logger.Fatal("Failed to authenticate to IMAP server", zap.Error(err))
	}
	logger.Info("Authenticated to IMAP server", zap.String("user", koboMailConfig.IMAPConfig.IMAPUser))
	defer imapConnection.Logout()

	// Select mailbox so we can search on it
	mbox, err := imapConnection.SelectMailbox(koboMailConfig.IMAPConfig.IMAPFolder)
	if err != nil {
		nickeldbus.DialogAddOKButton()
		nickeldbus.DialogUpdate("Failed to select IMAP mailbox " + koboMailConfig.IMAPConfig.IMAPFolder + ": " + err.Error())

		logger.Fatal("Failed to select IMAP mailbox", zap.Error(err))
	}
	logger.Info("IMAP mailbox selected", zap.String("name", mbox.Name))

	// Apply the search criteria and check if there's any emails with that criteria
	if koboMailConfig.IMAPConfig.EmailUnseen == "true" {
		imapConnection.SearchCriteria.WithoutFlags = []string{"\\Seen"}
	}
	if koboMailConfig.IMAPConfig.EmailFlagType == config.EmailFlagTypePlus {
		criterium := strings.Replace(koboMailConfig.IMAPConfig.IMAPUser, "@", "+"+koboMailConfig.IMAPConfig.EmailFlag+"@", 1)
		imapConnection.SearchCriteria.Header.Add("TO", criterium)
	} else if koboMailConfig.IMAPConfig.EmailFlagType == config.EmailFlagTypeSubject {
		criterium := koboMailConfig.IMAPConfig.EmailFlag
		imapConnection.SearchCriteria.Header.Add("SUBJECT", criterium)
	}

	messages, err := imapConnection.CollectMessages()
	if err != nil {
		nickeldbus.DialogAddOKButton()
		nickeldbus.DialogUpdate("Failed to fetch messages: " + err.Error())

		logger.Fatal("Failed to fetch messages", zap.Error(err))
	}
	numberOfEmailsFound := len(messages)
	logger.Info("Fetched emails", zap.Int("number_of_emails_found", numberOfEmailsFound))

	// Process the collected messages
	if numberOfEmailsFound > 0 {
		nickeldbus.DialogUpdate("Found " + strconv.Itoa(numberOfEmailsFound) + " emails to process. Please wait...")
	} else {
		nickeldbus.DialogAddOKButton()
		nickeldbus.DialogUpdate("No emails found, nothing to be done.")
		os.Exit(0)
	}

	numberOfEbooksProcessed := 0

	for _, msg := range messages {
		err := msg.FetchDetails()
		if err != nil {
			nickeldbus.DialogAddOKButton()
			nickeldbus.DialogUpdate("Exiting, failed to process message: " + err.Error())
			logger.Fatal("Failed to process message", zap.Any("message", msg), zap.Error(err))
		}
		logger.Info("Processing message", zap.Any("message", msg))

		downloadedAttachmentsCount, err := msg.ProcessAttachments(koboMailConfig.ProcessingConfig.Filetypes, config.DefaultLibraryPath)
		if err != nil {
			nickeldbus.DialogAddOKButton()
			nickeldbus.DialogUpdate("Exiting, failed to process attachment: " + err.Error())
			logger.Fatal("Failed to process attachment", zap.Error(err))
		}

		numberOfEbooksProcessed = numberOfEbooksProcessed + downloadedAttachmentsCount

		if koboMailConfig.IMAPConfig.EmailDelete == "true" {
			err := imapConnection.DeleteMessage(msg)
			if err != nil {
				nickeldbus.DialogAddOKButton()
				nickeldbus.DialogUpdate("Exiting, failed to delete message: " + err.Error())
				logger.Fatal("Failed to delete message", zap.Any("message", msg), zap.Error(err))
			}
		}
	}

	if nickeldbus.UseNickelDbus {
		if numberOfEbooksProcessed > 0 {
			// Rescan the library for the new ebooks
			_, err := nickeldbus.LibraryRescanFull("30000")
			if err != nil {
				logger.Error("Could not update library", zap.Error(err))
			}
			logger.Debug("Updated library")

			nickeldbus.DialogCreate("Processed " + strconv.Itoa(numberOfEbooksProcessed) + " new ebooks.")
			nickeldbus.DialogAddOKButton()
		} else {
			nickeldbus.DialogCreate("No emails found, nothing to be done.")
			nickeldbus.DialogAddOKButton()
		}
	} else {
		// After finishing loading all messages simulate the USB cable connect
		// but only if there were any messages processed, no need to bug the user if there was nothing new
		if numberOfEbooksProcessed > 0 {
			logger.Info("Simulating plugging USB cable and wait 10s for the user to click on the connect button")
			nickelUSBplugAddRemove("add")
			time.Sleep(10 * time.Second)
			logger.Info("Simulating unplugging USB cable")
			nickelUSBplugAddRemove("remove")
			// After this Nickel will do the job to import the new files loaded into the KoboMailLibrary folder
		}
	}
}
