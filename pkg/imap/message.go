// Package imap implements all IMAP interactions of KoboMail
package imap

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clisboa/kobomail/pkg/logger"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-message/mail"
	"go.uber.org/zap"
)

type message struct {
	imapMessage   *imap.Message
	messageReader *mail.Reader

	Date    time.Time
	Sender  string
	Subject string
}

func (msg *message) getMessageReader() (*mail.Reader, error) {
	if msg.messageReader != nil {
		return msg.messageReader, nil
	}

	msgBody := msg.imapMessage.GetBody(&imap.BodySectionName{})
	if msgBody == nil {
		return nil, fmt.Errorf("server did not return message body")
	}

	msgReader, err := mail.CreateReader(msgBody)
	if err != nil {
		return nil, err
	}

	msg.messageReader = msgReader
	return msgReader, nil
}

func (msg *message) FetchDetails() error {
	msgReader, err := msg.getMessageReader()
	if err != nil {
		return err
	}
	messageDate, _ := msgReader.Header.Date()
	messageSender, _ := msgReader.Header.AddressList("from")
	messageSubject, _ := msgReader.Header.Subject()

	msg.Date = messageDate
	msg.Sender = messageSender[0].Address
	msg.Subject = messageSubject

	return nil
}

func (msg *message) ProcessAttachments(allowedExtensions []string, destinationPath string) (int, error) {
	msgReader, err := msg.getMessageReader()
	if err != nil {
		return 0, err
	}

	downloadedAttachmentCount := 0

	// Process each message part, there might be multiple attachments
	for {
		p, err := msgReader.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		switch h := p.Header.(type) {
		// This is an attachment
		case *mail.AttachmentHeader:
			attachmentFileName, _ := h.Filename()
			attachmentFileExtension := strings.Trim(filepath.Ext(attachmentFileName), ".")

			// Only save the attachment if the filetype is allowed
			if containsFiletype(allowedExtensions, attachmentFileExtension) {
				// Check if the file is a kepub, rename it to .kepub.epub so kobo can properly handle it
				if attachmentFileExtension == "kepub" {
					attachmentFileName += ".epub"
				}

				logger.Debug("Downloading attachment", zap.String("filename", attachmentFileName))

				attachmentContent, _ := io.ReadAll(p.Body)
				// Write the whole body at once
				err = os.WriteFile(destinationPath+"/"+attachmentFileName, attachmentContent, 0644)
				if err != nil {
					return 0, err
				}
				logger.Info("Succesfully downloaded attachment", zap.String("filename", attachmentFileName))
				downloadedAttachmentCount++
			}
		}
	}

	return downloadedAttachmentCount, nil
}

func containsFiletype(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
