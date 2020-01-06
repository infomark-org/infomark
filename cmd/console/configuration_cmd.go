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

package console

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/franela/goblin"
	"github.com/infomark-org/infomark/auth"
	"github.com/infomark-org/infomark/configuration"
	"github.com/infomark-org/infomark/configuration/bytefmt"
	"github.com/infomark-org/infomark/configuration/fs"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	// "os"
	"text/template"
	"time"

	redis "github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ConfigurationCmd starts the infomark configuration
var ConfigurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "infomark configuration commands",
}

func init() {
	ConfigurationCmd.AddCommand(CreateConfiguration)
	ConfigurationCmd.AddCommand(TestConfiguration)
	ConfigurationCmd.AddCommand(CreateDockercompose)

}

func DurationFromString(dur string) time.Duration {
	d, err := time.ParseDuration(dur)
	if err != nil {
		panic(err)
	}
	return d
}

func ByteFromString(str string) bytefmt.ByteSize {
	d, err := bytefmt.FromString(str)
	if err != nil {
		panic(err)
	}
	return d
}

func GenerateExampleConfiguration(domain string, root_path string) *configuration.ConfigurationSchema {
	config := &configuration.ConfigurationSchema{}

	config.Server.Version = 1

	config.Server.HTTP.UseHTTPS = false
	config.Server.HTTP.Port = 2020
	config.Server.HTTP.Domain = domain
	config.Server.HTTP.Timeouts.Read = DurationFromString("30s")
	config.Server.HTTP.Timeouts.Write = DurationFromString("30s")
	config.Server.HTTP.Limits.MaxHeader = 1 * bytefmt.Megabyte
	config.Server.HTTP.Limits.MaxRequestJSON = 2 * bytefmt.Megabyte
	config.Server.HTTP.Limits.MaxAvatar = 1 * bytefmt.Megabyte
	config.Server.HTTP.Limits.MaxSubmission = 4 * bytefmt.Megabyte

	config.Server.Debugging.Enabled = false
	config.Server.Debugging.LoginID = int64(1)
	config.Server.Debugging.LoginIsRoot = false
	config.Server.Debugging.LogLevel = "debug"
	config.Server.Debugging.Fixtures = root_path + "/fixtures"

	config.Server.DistributeJobs = true

	config.Server.Authentication.JWT.Secret = auth.GenerateToken(32)
	config.Server.Authentication.JWT.AccessExpiry = 15 * time.Minute
	config.Server.Authentication.JWT.RefreshExpiry = DurationFromString("10h")
	config.Server.Authentication.Session.Secret = auth.GenerateToken(32)
	config.Server.Authentication.Session.Cookies.Secure = config.Server.HTTP.UseHTTPS
	config.Server.Authentication.Session.Cookies.Lifetime = DurationFromString("24h")
	config.Server.Authentication.Session.Cookies.IdleTimeout = DurationFromString("60m")
	config.Server.Authentication.Password.MinLength = 7

	config.Server.Authentication.TotalRequestsPerMinute = 100
	config.Server.Cronjobs.ZipSubmissionsIntervall = DurationFromString("5m")

	config.Server.Email.Send = false
	config.Server.Email.SendmailBinary = "/usr/sbin/sendmail"
	config.Server.Email.From = fmt.Sprintf("no-reply@%s", config.Server.HTTP.Domain)
	config.Server.Email.ChannelSize = 300

	config.Server.Services.Redis.Host = "localhost"
	config.Server.Services.Redis.Port = 6379
	config.Server.Services.Redis.Database = 0

	config.Server.Services.RabbitMQ.Host = "localhost"
	config.Server.Services.RabbitMQ.Port = 5672
	config.Server.Services.RabbitMQ.User = "rabbitmq_user"
	config.Server.Services.RabbitMQ.Password = auth.GenerateToken(32)
	config.Server.Services.RabbitMQ.Key = "rabbitmq_key"

	config.Server.Services.Postgres.Host = "localhost"
	config.Server.Services.Postgres.Port = 5432
	config.Server.Services.Postgres.User = "database_user"
	config.Server.Services.Postgres.Database = "infomark"
	config.Server.Services.Postgres.Password = auth.GenerateToken(32)

	config.Server.Paths.Uploads = root_path + "/uploads"
	config.Server.Paths.Common = root_path + "/common"
	config.Server.Paths.GeneratedFiles = root_path + "/generated_files"

	config.Worker.Version = config.Server.Version
	config.Worker.Services.RabbitMQ = config.Server.Services.RabbitMQ
	config.Worker.Workdir = "/tmp"
	config.Worker.Void = false
	config.Worker.Docker.MaxMemory = 500 * bytefmt.Megabyte
	config.Worker.Docker.Timeout = 5 * time.Second
	return config
}

var CreateConfiguration = &cobra.Command{
	Use:   "create",
	Short: "will create and print a configuration to stdout",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		ex, err := os.Executable()
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		exPath := path.Join(filepath.Dir(ex), "files")

		config := GenerateExampleConfiguration("localhost", exPath)
		serialized_config, err := yaml.Marshal(&config)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Println(string(serialized_config))
	},
}

func showResult(report goblin.DetailedReporter, err error, text string) {
	if err != nil {
		report.ItFailed(text)
		report.BeginDescribe(err.Error())
		report.EndDescribe()
	} else {
		report.ItPassed(text)
	}
}

var TestConfiguration = &cobra.Command{
	Use:   "test [configfile]",
	Short: "will create and print a configuration to stdout",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		status_code := 0

		report := goblin.DetailedReporter{}
		report.SetTextFancier(&goblin.TerminalFancier{})
		report.BeginDescribe("Test configuration services")

		config, err := configuration.ParseConfiguration(args[0])
		showResult(report, err, "read configuration from file")
		if err != nil {
			status_code = -1
		}

		// Try to connect to postgres
		db, err := sqlx.Connect("postgres", config.Server.PostgresURL())
		showResult(report, err, "connect to postgres db")
		if err != nil {
			status_code = -1
		} else {
			db.Close()
		}

		// Try to connect to Redis
		option, _ := redis.ParseURL(config.Server.RedisURL())

		redisClient := redis.NewClient(option)
		_, err = redisClient.Ping().Result()
		showResult(report, err, "connect to redis url")
		if err != nil {
			status_code = -1
		} else {
			redisClient.Close()
		}

		report.EndDescribe()

		report.BeginDescribe("Test configuration paths")
		err = fs.DirExists(config.Server.Paths.Common)
		showResult(report, err, "common path readable")
		if err != nil {
			status_code = -1
		}

		err = fs.IsDirWriteable(config.Server.Paths.Uploads)
		showResult(report, err, "upload path writeable")
		if err != nil {
			status_code = -1
		}

		err = fs.IsDirWriteable(config.Server.Paths.GeneratedFiles)
		showResult(report, err, "generated_files path writeable")
		if err != nil {
			status_code = -1
		}

		privacyFile := fmt.Sprintf("%s/privacy_statement.md", config.Server.Paths.Common)
		err = fs.FileExists(privacyFile)
		showResult(report, err, fmt.Sprintf("Read privacy Statement from %s", privacyFile))
		if err != nil {
			status_code = -1
		}
		report.EndDescribe()

		os.Exit(status_code)
	},
}

var CreateDockercompose = &cobra.Command{
	Use:   "create-compose [configfile]",
	Short: "create docker-compose file from config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := configuration.ParseConfiguration(args[0])
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		docker_compose, err := template.New("docker-compose").Parse(
			`version: "3"
services:
  rabbitmq_host:
    image: rabbitmq:3.7.3-management-alpine
    environment:
      - RABBITMQ_DEFAULT_USER={{.Server.Services.RabbitMQ.User}}
      - RABBITMQ_DEFAULT_PASS={{.Server.Services.RabbitMQ.Password}}
    ports:
      - 127.0.0.1:{{.Server.Services.RabbitMQ.Port}}:5672
      - 127.0.0.1:15672:15672
    volumes:
      - rabbitmq_volume:/data
  postgres_host:
    image: postgres:11.2-alpine
    environment:
      - POSTGRES_DB={{.Server.Services.Postgres.Database}}
      - POSTGRES_USER={{.Server.Services.Postgres.User}}
      - POSTGRES_PASSWORD={{.Server.Services.Postgres.Password}}
      - PGPASSWORD={{.Server.Services.Postgres.Password}}
    ports:
      - 127.0.0.1:{{.Server.Services.Postgres.Port}}:5432
    volumes:
      - postgres_volume:/var/lib/postgresql/data
  redis_host:
    image: redis:5.0.4-alpine
    ports:
      - 127.0.0.1:{{.Server.Services.Redis.Port}}:6379
volumes:
  rabbitmq_volume:
  postgres_volume:`)

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		var tpl bytes.Buffer
		err = docker_compose.Execute(&tpl, config)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		fmt.Println(tpl.String())
	},
}
