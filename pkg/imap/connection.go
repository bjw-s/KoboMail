// Package imap implements all IMAP interactions of KoboMail
package imap

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// Connection is a simple implementation of an IMAP connection
type Connection struct {
	host   string
	port   int
	client *client.Client

	SearchCriteria *imap.SearchCriteria
}

func (ic *Connection) connect() error {
	tlsn := ""
	connStr := fmt.Sprintf("%s:%v", ic.host, ic.port)

	tlsc := &tls.Config{}
	if tlsn != "" {
		tlsc.ServerName = tlsn
	}

	numRetries := 3
	c, err := client.DialTLS(connStr, tlsc)
	if err != nil {
		for numRetries > 0 {
			time.Sleep(1 * time.Second)
			c, err = client.DialTLS(connStr, tlsc)
			if err != nil {
				numRetries--
			} else {
				break
			}
		}

		if err != nil {
			return err
		}
	}

	ic.client = c
	ic.SearchCriteria = imap.NewSearchCriteria()
	return nil
}

// ConnectToServer instantiates a new connection to an IMAP server
func ConnectToServer(host string, port int) (*Connection, error) {
	connection := Connection{
		host: host,
		port: port,
	}

	err := connection.connect()
	if err != nil {
		return nil, err
	}
	return &connection, nil
}

// Login identifies the client to the server and carries the plaintext password
// authenticating this user.
func (ic *Connection) Login(username string, password string) error {
	return ic.client.Login(username, password)
}

// Logout gracefully closes the connection.
func (ic *Connection) Logout() error {
	return ic.client.Logout()
}

// SelectMailbox selects a mailbox so that messages in the mailbox can be accessed.
func (ic *Connection) SelectMailbox(mailbox string) (*imap.MailboxStatus, error) {
	return ic.client.Select(mailbox, false)
}

// CollectMessages collects the messages based on the criteria set on the IMAPConnection.
func (ic *Connection) CollectMessages() ([]*message, error) {
	uids, err := ic.client.Search(ic.SearchCriteria)
	if err != nil {
		return nil, err
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	// Fetch the emails list
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}
	messages := make(chan *imap.Message)
	done := make(chan error, 1)
	go func() {
		done <- ic.client.Fetch(seqset, items, messages)
	}()

	var collectedMessages []*message
	for msg := range messages {
		if msg != nil {
			collectedMessage := message{
				imapMessage: msg,
			}

			collectedMessages = append(collectedMessages, &collectedMessage)
		}
	}

	return collectedMessages, nil
}

// DeleteMessage deletes the specified message on the server.
func (ic *Connection) DeleteMessage(msg *message) error {
	if msg != nil {
		seqset := new(imap.SeqSet)
		seqset.AddNum(msg.imapMessage.SeqNum)

		done := make(chan error, 1)
		go func() {
			done <- ic.client.Store(seqset, imap.AddFlags, []interface{}{imap.DeletedFlag}, nil)
		}()
		err := <-done
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
