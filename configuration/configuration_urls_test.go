// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2020-present InfoMark.org
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

package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/franela/goblin"
)

// standardHTTPPort and standardHTTPSPort are the well-known ports that the URL
// builder omits from the rendered URL.
const (
	standardHTTPPort  = 80
	standardHTTPSPort = 443
	customPort        = 8080
)

func TestConfigurationURLs(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("ServerConfigurationSchema.URL", func() {

		// On plain HTTP with the standard port 80 the port must be omitted.
		g.It("Should omit port 80 for http", func() {
			config := &ServerConfigurationSchema{}
			config.HTTP.UseHTTPS = false
			config.HTTP.Domain = "example.org"
			config.HTTP.Port = standardHTTPPort
			g.Assert(config.URL()).Equal("http://example.org")
		})

		// On HTTPS with the standard port 443 the port must be omitted.
		g.It("Should omit port 443 for https", func() {
			config := &ServerConfigurationSchema{}
			config.HTTP.UseHTTPS = true
			config.HTTP.Domain = "example.org"
			config.HTTP.Port = standardHTTPSPort
			g.Assert(config.URL()).Equal("https://example.org")
		})

		// A non-standard port must be rendered explicitly for http.
		g.It("Should include a custom port for http", func() {
			config := &ServerConfigurationSchema{}
			config.HTTP.UseHTTPS = false
			config.HTTP.Domain = "example.org"
			config.HTTP.Port = customPort
			g.Assert(config.URL()).Equal("http://example.org:8080")
		})

		// URL() only omits the port when it is the default for the active
		// protocol (80 for http, 443 for https). With UseHTTPS false and port
		// 443, the protocol is http but 443 is not http's default port, so the
		// port MUST be rendered: "http://example.org" would otherwise be
		// interpreted as port 80 and point at the wrong endpoint. Emitting
		// "http://example.org:443" is therefore the correct, unambiguous URL for
		// plain http served on port 443.
		g.It("Should include port 443 when https is disabled", func() {
			config := &ServerConfigurationSchema{}
			config.HTTP.UseHTTPS = false
			config.HTTP.Domain = "example.org"
			config.HTTP.Port = standardHTTPSPort
			g.Assert(config.URL()).Equal("http://example.org:443")
		})
	})

	g.Describe("ServerConfigurationSchema.ExternalURL", func() {

		// ExternalURL always drops the port and reflects the protocol.
		g.It("Should render an https external url without the port", func() {
			config := &ServerConfigurationSchema{}
			config.HTTP.UseHTTPS = true
			config.HTTP.Domain = "example.org"
			config.HTTP.Port = customPort
			g.Assert(config.ExternalURL()).Equal("https://example.org")
		})

		g.It("Should render an http external url without the port", func() {
			config := &ServerConfigurationSchema{}
			config.HTTP.UseHTTPS = false
			config.HTTP.Domain = "example.org"
			config.HTTP.Port = customPort
			g.Assert(config.ExternalURL()).Equal("http://example.org")
		})
	})

	g.Describe("ServerConfigurationSchema.HTTPAddr", func() {

		// HTTPAddr must render as a bind address of the form ":port".
		g.It("Should render the bind address from the port", func() {
			config := &ServerConfigurationSchema{}
			config.HTTP.Port = customPort
			g.Assert(config.HTTPAddr()).Equal(":8080")
		})
	})

	g.Describe("ServerConfigurationSchema.SendEmail", func() {

		// Sending is enabled only when both the flag is set and a binary is given.
		g.It("Should enable sending only with flag and binary", func() {
			config := &ServerConfigurationSchema{}
			config.Email.Send = true
			config.Email.SendmailBinary = "/usr/sbin/sendmail"
			g.Assert(config.SendEmail()).IsTrue()
		})

		// A missing binary must disable sending even when the flag is set.
		g.It("Should disable sending when the binary is empty", func() {
			config := &ServerConfigurationSchema{}
			config.Email.Send = true
			config.Email.SendmailBinary = ""
			g.Assert(config.SendEmail()).IsFalse()
		})

		// The flag being false must disable sending regardless of the binary.
		g.It("Should disable sending when the flag is off", func() {
			config := &ServerConfigurationSchema{}
			config.Email.Send = false
			config.Email.SendmailBinary = "/usr/sbin/sendmail"
			g.Assert(config.SendEmail()).IsFalse()
		})
	})

	g.Describe("RabbitMQConfiguration.URL", func() {

		// The AMQP URL must embed user, password, host and port in order.
		g.It("Should build an amqp url", func() {
			config := &RabbitMQConfiguration{
				Host:     "broker",
				Port:     5672,
				User:     "guest",
				Password: "secret",
				Key:      "ignored",
			}
			g.Assert(config.URL()).Equal("amqp://guest:secret@broker:5672/")
		})
	})

	g.Describe("ParseConfiguration errors", func() {

		// A missing configuration file must surface as an error rather than a panic.
		g.It("Should error on a missing file", func() {
			config, err := ParseConfiguration(filepath.Join(os.TempDir(), "does-not-exist-infomark.yml"))
			g.Assert(err == nil).IsFalse()
			g.Assert(config == nil).IsTrue()
		})

		// Malformed YAML must produce a parse error. We write a temporary file
		// containing content that cannot be a valid mapping for the schema.
		g.It("Should error on malformed yaml", func() {
			temporaryFile, err := os.CreateTemp("", "infomark-malformed-*.yml")
			g.Assert(err).Equal(nil)
			defer os.Remove(temporaryFile.Name())

			// A bare scalar cannot unmarshal into the configuration mapping.
			_, err = temporaryFile.WriteString("::: not : valid : yaml :::\n\t- broken\n")
			g.Assert(err).Equal(nil)
			g.Assert(temporaryFile.Close()).Equal(nil)

			_, parseErr := ParseConfiguration(temporaryFile.Name())
			g.Assert(parseErr == nil).IsFalse()
		})
	})
}
