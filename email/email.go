// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package email

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/spf13/viper"
)

// Email contains all information to use sendmail
type Email struct {
	From    string
	To      string
	Subject string
	Body    string
}

// OutgoingEmailsChannel is a light-weight go-routine to send emails
var OutgoingEmailsChannel chan *Email

// NewEmail creates a new email structure
func NewEmail(toEmail string, subject string, body string) *Email {
	email := &Email{
		From:    viper.GetString("email_from"),
		To:      toEmail,
		Subject: subject,
		Body:    body,
	}
	return email
}

// NewEmailFromUser creates a new email structure and appends the sender information
func NewEmailFromUser(toEmail string, subject string, body string, user *model.User) *Email {
	email := &Email{
		From:    viper.GetString("email_from"),
		To:      toEmail,
		Subject: subject,
		Body:    fmt.Sprintf("%s\n\n----------\nSender is %s\nSent via InfoMark\n", body, user.FullName()),
	}
	return email
}

// NewEmailFromTemplate creates a new email structure filling a template file
func NewEmailFromTemplate(toEmail string, subject string, file string, data map[string]string) (*Email, error) {
	body, err := LoadAndFillTemplate(file, data)
	if err != nil {
		return nil, err
	}
	return NewEmail(toEmail, subject, body), nil
}

// Emailer any object that can send
type Emailer interface {
	Send(e *Email) error
	// LoadTemplate(file string, data map[string]string) error
}

// SendMailer uses the sendmail binary to send emails.
type SendMailer struct {
	Binary string
}

// TerminalMailer prints the email to the terminal.
type TerminalMailer struct{}

// VoidMailer does nothing (to keep the unit test outputs clean)
type VoidMailer struct{}

// NewSendMailer creates an object that will send emails over sendmail
func NewSendMailer() *SendMailer {
	sendmailBinary := viper.GetString("sendmail_binary")
	return &SendMailer{Binary: sendmailBinary}
}

// NewTerminalMailer creates an object that printout the email in the terminal
func NewTerminalMailer() *TerminalMailer {
	return &TerminalMailer{}
}

// NewVoidMailer creates an object that drops any email
func NewVoidMailer() *VoidMailer {
	return &VoidMailer{}
}

// SendMail is ready-to-use instance for sendmail
var SendMail = NewSendMailer()

// TerminalMail is ready-to-use instance for displaying emails in the terminal
var TerminalMail = NewTerminalMailer()

// VoidMail is ready-to-use instance for dropping outgoing emails
var VoidMail = NewVoidMailer()

// DefaultMail is the default instance used by infomark
var DefaultMail Emailer

func init() {
	DefaultMail = TerminalMail
	OutgoingEmailsChannel = make(chan *Email, 300)
}

// Send will drop any outgoing email
func (sm *VoidMailer) Send(e *Email) error {
	return nil
}

// BackgroundSend will send emails enqueued in a channel
func BackgroundSend(emails <-chan *Email) {
	for email := range emails {
		DefaultMail.Send(email)
	}
}

// Send prints everything to stdout.
func (sm *TerminalMailer) Send(e *Email) error {
	fmt.Printf("From: %s\n", e.From)
	fmt.Printf("To: %s\n", e.To)
	fmt.Printf("Subject: %s\n", e.Subject)
	fmt.Printf("\n")
	fmt.Printf("%s", e.Body)
	return nil
}

// Send uses `sendmail` to deliver emails.
func (sm *SendMailer) Send(e *Email) error {

	cmd := exec.Command(sm.Binary, "-t")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	pw, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	pw.Write([]byte(fmt.Sprintf("From: %s\n", e.From)))
	pw.Write([]byte(fmt.Sprintf("To: %s\n", e.To)))
	pw.Write([]byte(fmt.Sprintf("Subject: %s\n", e.Subject)))
	pw.Write([]byte(fmt.Sprintf("\n"))) // blank line separating headers from body
	pw.Write([]byte(fmt.Sprintf("%s", e.Body)))

	err = pw.Close()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

// LoadAndFillTemplate loads a template and fills out the placeholders.
func LoadAndFillTemplate(file string, data map[string]string) (string, error) {
	rootDir := viper.GetString("email_templates_dir")
	t, err := template.ParseFiles(fmt.Sprintf("%s/%s", rootDir, file))
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	return tpl.String(), nil
}
