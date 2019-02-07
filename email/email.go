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

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"

	"github.com/spf13/viper"
)

func SendEmail(toEmail string, subject string, body string) error {
	// fromEmail := "no-reply@info2.informatik.uni-tuebingen.de" //viper.GetString("email_from")
	// sendmail_binary := "/usr/sbin/sendmail"                   // viper.GetString("email_sendmail")
	fromEmail := viper.GetString("email_from")
	sendmail_binary := viper.GetString("sendmail_binary")

	fmt.Println("from", fromEmail)
	fmt.Println("sendmail_binary", sendmail_binary)

	cmd := exec.Command(sendmail_binary, "-t")
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

	pw.Write([]byte(fmt.Sprintf("From: %s\n", fromEmail)))
	pw.Write([]byte(fmt.Sprintf("To: %s\n", toEmail)))
	pw.Write([]byte(fmt.Sprintf("Subject: %s\n", subject)))
	pw.Write([]byte(fmt.Sprintf("\n"))) // blank line separating headers from body
	pw.Write([]byte(fmt.Sprintf("%s", body)))
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

func ParseEmailTemplate(file string, data map[string]string) (string, error) {
	root_dir := viper.GetString("email_templates_dir")
	fmt.Println(root_dir)
	t, err := template.ParseFiles(fmt.Sprintf("%s/%s", root_dir, file))
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	return tpl.String(), nil
}

func main() {

	viper.SetConfigFile("/home/wieschol/git/github.com/cgtuebingen/infomark-go/infomark-backend/infomark-backend.yml")
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	body, err := ParseEmailTemplate(
		"request_password_token.de.txt",
		map[string]string{
			"first_name":  "Patrick",
			"last_name":   "Wiesch",
			"reset_url":   "http://info2.informatik.uni-tuebingen.de/reset",
			"reset_token": "sdjfgsdjkfddd",
		},
	)
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Println(body)
	if err == nil {
		SendEmail("patrick.wieschollek@uni-tuebingen.de", "GoSubject222", body)

	} else {
		fmt.Println(err)
	}
}
