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
	"testing"
	"time"

	"github.com/franela/goblin"
)

func TestConfiguration(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Configuration", func() {

		g.It("Should read configuration", func() {

			config, err := ParseConfiguration("example.yml")
			g.Assert(err).Equal(nil)
			g.Assert(config.Server.HTTP.Port).Equal(3000)
			g.Assert(config.Server.HTTP.Domain).Equal("sub.domain.com")

			g.Assert(config.Server.Authentication.Email.Verify).Equal(true)
			g.Assert(config.Server.Debugging.Enabled).Equal(false)
			g.Assert(config.Server.Debugging.LoginID).Equal(int64(1))
			g.Assert(config.Server.Debugging.LoginIsRoot).Equal(false)
			g.Assert(config.Server.Debugging.LogLevel).Equal("debug")

		})

		g.It("Should have correct intervall", func() {

			config := &ServerConfigurationSchema{}
			config.Cronjobs.ZipSubmissionsIntervall = 4 * time.Second
			g.Assert(config.CronjobsZipSubmissionsIntervall()).Equal("@ every 4s")

		})

		g.It("Should have correct postgres url", func() {

			config := &ServerConfigurationSchema{}
			config.Services.Postgres.User = "user"
			config.Services.Postgres.Password = "pass"
			config.Services.Postgres.Host = "1.2.3.4"
			config.Services.Postgres.Port = 5678
			config.Services.Postgres.Database = "dbname"

			g.Assert(config.PostgresURL()).Equal("postgres://user:pass@1.2.3.4:5678/dbname?sslmode=disable&connect_timeout=1")

		})

		g.It("Should have correct redis url", func() {

			config := &ServerConfigurationSchema{}
			config.Services.Redis.Host = "1.2.3.4"
			config.Services.Redis.Port = 5678
			config.Services.Redis.Database = 0

			g.Assert(config.RedisURL()).Equal("redis://1.2.3.4:5678/0")

		})

	})

}
