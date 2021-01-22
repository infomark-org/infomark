// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
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

	"github.com/infomark-org/infomark/model"
)

// Email contains all information to use sendmail
type Email struct {
	From    string
	ReplyTo string
	To      string
	Subject string
	Body    string
}

// OutgoingEmailsChannel is a light-weight go-routine to send emails
var OutgoingEmailsChannel chan *Email

// NewEmail creates a new email structure
func NewEmail(from string, toEmail string, subject string, body string) *Email {
	email := &Email{
		From:    from,
		To:      toEmail,
		Subject: subject,
		Body:    body,
	}
	return email
}

// NewEmailFromUser creates a new email structure and appends the sender information
func NewEmailFromUser(from string, toEmail string, subject string, body string, user *model.User) *Email {
	email := &Email{
		From:    from,
		ReplyTo: user.Email,
		To:      toEmail,
		Subject: subject,
		Body:    fmt.Sprintf("%s\n\n----------\nSender is %s\nSent via InfoMark\n", body, user.FullName()),
	}
	return email
}

// NewEmailFromTemplate creates a new email structure filling a template file
func NewEmailFromTemplate(from string, toEmail string, subject string, tpl *template.Template, data map[string]string) (*Email, error) {
	body, err := FillTemplate(tpl, data)
	if err != nil {
		return nil, err
	}
	return NewEmail(from, toEmail, subject, body), nil
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
func NewSendMailer(sendmailBinary string) *SendMailer {
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
var SendMail *SendMailer

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
	if e.ReplyTo != "" {
		fmt.Printf("Reply-To: %s\n", e.ReplyTo)
	}
	fmt.Printf("Subject: %s\n", e.Subject)
	fmt.Printf("Content-Type: text/plain; charset=\"utf-8\"\n")
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
	if e.ReplyTo != "" {
		pw.Write([]byte(fmt.Sprintf("Reply-To: %s\n", e.ReplyTo)))
	}
	pw.Write([]byte(fmt.Sprintf("Subject: %s\n", e.Subject)))
	pw.Write([]byte("Content-Type: text/plain; charset=\"utf-8\"\n"))
	pw.Write([]byte("\n")) // blank line separating headers from body
	pw.Write([]byte(e.Body))

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

// FillTemplate loads a template and fills out the placeholders.
func FillTemplate(t *template.Template, data map[string]string) (string, error) {
	var tpl bytes.Buffer
	err := t.Execute(&tpl, data)
	return tpl.String(), err
}

const (
	confirmEmailTemplateSrcEN = `Hi {{.first_name}} {{.last_name}}!

You must now confirm your email address to:
   - Log into our system and upload your homework solutions
   - Reset your password
   - Receive account alerts

Please use the following link to confirm your email address:

{{.confirm_email_url}}/{{.confirm_email_address}}/{{.confirm_email_token}}

`

	requestPasswordTokenTemailTemplateSrcEN = `Hi {{.first_name}} {{.last_name}}!

We got a request to change your password. You can change your password using the following link.

{{.reset_password_url}}/{{.email_address}}/{{.reset_password_token}}

If you have not requested the change, you can ignore this mail.

Your password can only be changed manually by you.

`
)

var ConfirmEmailTemplateEN *template.Template = template.Must(template.New("confirmEmailTemplateSrcEN").Parse(confirmEmailTemplateSrcEN))
var RequestPasswordTokenTemailTemplateEN *template.Template = template.Must(template.New("requestPasswordTokenTemailTemplateSrcEN").Parse(requestPasswordTokenTemailTemplateSrcEN))
