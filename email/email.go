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

	"github.com/spf13/viper"
)

type Email struct {
	From    string
	To      string
	Subject string
	Body    string
}

func NewEmail(toEmail string, subject string, body string) *Email {
	email := &Email{
		From:    viper.GetString("email_from"),
		To:      toEmail,
		Subject: subject,
		Body:    body,
	}
	return email
}

func NewEmailFromTemplate(toEmail string, subject string, file string, data map[string]string) (*Email, error) {
	body, err := LoadAndFillTemplate(file, data)
	if err != nil {
		return nil, err
	}
	return NewEmail(toEmail, subject, body), nil
}

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

func NewSendMailer() *SendMailer {
	sendmail_binary := viper.GetString("sendmail_binary")
	return &SendMailer{Binary: sendmail_binary}
}

func NewTerminalMailer() *TerminalMailer {
	return &TerminalMailer{}
}

func NewVoidMailer() *VoidMailer {
	return &VoidMailer{}
}

var SendMail = NewSendMailer()
var TerminalMail = NewTerminalMailer()
var VoidMail = NewVoidMailer()
var DefaultMail Emailer

func init() {
	DefaultMail = TerminalMail
}

func (sm *VoidMailer) Send(e *Email) error {
	return nil
}

// TerminalMailer prints everything to stdout.
func (sm *TerminalMailer) Send(e *Email) error {
	fmt.Printf("From: %s\n", e.From)
	fmt.Printf("To: %s\n", e.To)
	fmt.Printf("Subject: %s\n", e.Subject)
	fmt.Printf("\n")
	fmt.Printf("%s", e.Body)
	return nil
}

// SendMailer uses `sendmail` to deliver emails.
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
	root_dir := viper.GetString("email_templates_dir")
	// fmt.Println(root_dir)
	t, err := template.ParseFiles(fmt.Sprintf("%s/%s", root_dir, file))
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	return tpl.String(), nil
}

// func main() {

// 	viper.SetConfigFile("/home/wieschol/git/github.com/cgtuebingen/infomark-go/infomark-backend/infomark-backend.yml")
// 	if err := viper.ReadInConfig(); err == nil {
// 		fmt.Println("Using config file:", viper.ConfigFileUsed())
// 	}

// 	body, err := LoadAndFillTemplate(
// 		"request_password_token.de.txt",
// 		map[string]string{
// 			"first_name":  "Patrick",
// 			"last_name":   "Wiesch",
// 			"reset_url":   "http://info2.informatik.uni-tuebingen.de/reset",
// 			"reset_token": "sdjfgsdjkfddd",
// 		},
// 	)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	// fmt.Println(body)
// 	if err == nil {
// 		SendEmail("patrick.wieschollek@uni-tuebingen.de", "GoSubject222", body)

// 	} else {
// 		fmt.Println(err)
// 	}
// }
