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
	"html/template"
	"strings"
	"testing"

	"github.com/franela/goblin"
	"github.com/infomark-org/infomark/model"
)

func TestEmail(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("NewEmail", func() {

		// The constructor must copy each field verbatim and leave ReplyTo empty.
		g.It("Should populate the basic fields", func() {
			email := NewEmail("from@example.org", "to@example.org", "Subject line", "the body")
			g.Assert(email.From).Equal("from@example.org")
			g.Assert(email.To).Equal("to@example.org")
			g.Assert(email.Subject).Equal("Subject line")
			g.Assert(email.Body).Equal("the body")
			g.Assert(email.ReplyTo).Equal("")
		})
	})

	g.Describe("NewEmailFromUser", func() {

		// The user variant must set ReplyTo to the sender's email and append a
		// signature that contains the sender's full name.
		g.It("Should set reply-to and append the sender signature", func() {
			sender := &model.User{FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.org"}
			email := NewEmailFromUser("noreply@example.org", "to@example.org", "Hi", "message body", sender)
			g.Assert(email.ReplyTo).Equal("ada@example.org")
			// The body must retain the original message.
			g.Assert(strings.Contains(email.Body, "message body")).IsTrue()
			// The appended signature must carry the sender's full name.
			g.Assert(strings.Contains(email.Body, "Ada Lovelace")).IsTrue()
		})
	})

	g.Describe("FillTemplate", func() {

		// A template placeholder must be replaced by the matching data value.
		g.It("Should substitute placeholders from the data map", func() {
			parsed := template.Must(template.New("greeting").Parse("Hello {{.name}}!"))
			rendered, err := FillTemplate(parsed, map[string]string{"name": "World"})
			g.Assert(err).Equal(nil)
			g.Assert(rendered).Equal("Hello World!")
		})

		// A missing key renders as the empty string for html/template maps, so
		// the surrounding literal text must still be present.
		g.It("Should render a missing key as empty text", func() {
			parsed := template.Must(template.New("greeting").Parse("Hello [{{.name}}]"))
			rendered, err := FillTemplate(parsed, map[string]string{})
			g.Assert(err).Equal(nil)
			g.Assert(rendered).Equal("Hello []")
		})
	})

	g.Describe("NewEmailFromTemplate", func() {

		// Rendering the confirmation template must inject every provided value
		// into the resulting email body.
		g.It("Should render the confirm-email template into the body", func() {
			data := map[string]string{
				"first_name":            "Ada",
				"last_name":             "Lovelace",
				"confirm_email_url":     "https://example.org/confirm",
				"confirm_email_address": "ada@example.org",
				"confirm_email_token":   "token-123",
			}
			email, err := NewEmailFromTemplate(
				"noreply@example.org",
				"ada@example.org",
				"Confirm your email",
				ConfirmEmailTemplateEN,
				data,
			)
			g.Assert(err).Equal(nil)
			g.Assert(email.Subject).Equal("Confirm your email")
			g.Assert(strings.Contains(email.Body, "Ada Lovelace")).IsTrue()
			g.Assert(strings.Contains(email.Body, "https://example.org/confirm")).IsTrue()
			g.Assert(strings.Contains(email.Body, "ada@example.org")).IsTrue()
			g.Assert(strings.Contains(email.Body, "token-123")).IsTrue()
		})
	})

	g.Describe("VoidMailer", func() {

		// The void mailer must silently accept any email and never error, which
		// keeps unit-test output clean.
		g.It("Should drop emails without error", func() {
			mailer := NewVoidMailer()
			err := mailer.Send(NewEmail("a@b.c", "d@e.f", "s", "b"))
			g.Assert(err).Equal(nil)
		})
	})
}
