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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/infomark-org/infomark/configuration/bytefmt"
	"gopkg.in/yaml.v2"
)

type RabbitMQConfiguration struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Key      string `yaml:"key"`
}

func (config *RabbitMQConfiguration) RabbitMQURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%v/", config.User, config.Password, config.Host, config.Port)
}

type AuthenticationConfiguration struct {
	JWT struct {
		Secret        string        `yaml:"secret"`
		AccessExpiry  time.Duration `yaml:"access_expiry"`
		RefreshExpiry time.Duration `yaml:"refresh_expiry"`
	} `yaml:"jwt"`
	Session struct {
		Secret  string `yaml:"secret"`
		Cookies struct {
			Secure      bool          `yaml:"secure"`
			Lifetime    time.Duration `yaml:"lifetime"`
			IdleTimeout time.Duration `yaml:"idle_timeout"`
		} `yaml:"cookies"`
	} `yaml:"session"`
	Password struct {
		MinLength int `yaml:"min_length"`
	} `yaml:"password"`
	TotalRequestsPerMinute int64 `yaml:"total_requests_per_minute"`
}

func (config *ServerConfigurationSchema) URL() string {
	// TODO(patwie): When hosted in a sub-path, this will not work.
	//  In this case, consider to add an URL field.
	protocoll := "http"
	if config.HTTP.UseHTTPS {
		protocoll = "https"
	}
	return fmt.Sprintf("%s://%v", protocoll, config.HTTP.Domain)
}

type PathsConfiguration struct {
	Uploads        string `yaml:"uploads"`
	Common         string `yaml:"common"`
	GeneratedFiles string `yaml:"generated_files"`
}

type ServerConfigurationSchema struct {
	Version   int `json:"version"`
	Debugging struct {
		Enabled     bool   `yaml:"enabled"`
		LoginID     int64  `yaml:"login_id"`
		LoginIsRoot bool   `yaml:"login_is_root"`
		LogLevel    string `yaml:"log_level"`
		Fixtures    string `yaml:"fixtures"`
	} `yaml:"debugging"`
	HTTP struct {
		UseHTTPS bool `yaml:"use_https"`
		// Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Domain   string `yaml:"domain"`
		Timeouts struct {
			Read  time.Duration `yaml:"read"`
			Write time.Duration `yaml:"write"`
		} `yaml:"timeouts"`
		Limits struct {
			MaxHeader      bytefmt.ByteSize `yaml:"max_header"`
			MaxRequestJSON bytefmt.ByteSize `yaml:"max_request_json"`
			MaxAvatar      bytefmt.ByteSize `yaml:"max_avatar"`
			MaxSubmission  bytefmt.ByteSize `yaml:"max_submission"`
		} `yaml:"limits"`
	} `yaml:"http"`
	DistributeJobs bool                        `yaml:"distribute_jobs"`
	Authentication AuthenticationConfiguration `yaml:"authentication"`
	Cronjobs       struct {
		ZipSubmissionsIntervall time.Duration `yaml:"zip_submissions_intervall"`
	} `yaml:"cronjobs"`
	Email struct {
		Send           bool   `yaml:"send"`
		SendmailBinary string `yaml:"sendmail_binary"`
		From           string `yaml:"from"`
		ChannelSize    int    `yaml:"channel_size"`
	} `yaml:"email"`
	Services struct {
		Redis struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Database int    `yaml:"database"`
			// Connection string `yaml:"connection"`
		} `yaml:"redis"`
		RabbitMQ RabbitMQConfiguration `yaml:"rabbit_mq"`
		Postgres struct {
			// Connection string `yaml:"connection"`
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Database string `yaml:"database"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			Debug    bool   `yaml:"debug"`
		} `yaml:"database"`
	} `yaml:"services"`
	Paths PathsConfiguration `yaml:"paths"`
}

func (config *ServerConfigurationSchema) SendEmail() bool {
	return (config.Email.Send && config.Email.SendmailBinary != "")
}

func (config *ServerConfigurationSchema) PostgresURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable&connect_timeout=1",
		config.Services.Postgres.User,
		config.Services.Postgres.Password,
		config.Services.Postgres.Host,
		config.Services.Postgres.Port,
		config.Services.Postgres.Database,
	)
}

func (config *ServerConfigurationSchema) RedisURL() string {
	return fmt.Sprintf("redis://%s:%v/%v",
		config.Services.Redis.Host,
		config.Services.Redis.Port,
		config.Services.Redis.Database,
	)
}

func (config *ServerConfigurationSchema) HTTPAddr() string {
	return fmt.Sprintf(":%v", config.HTTP.Port)
}
func (config *ServerConfigurationSchema) CronjobsZipSubmissionsIntervall() string {
	secs := config.Cronjobs.ZipSubmissionsIntervall
	return fmt.Sprintf("@ every %s", secs)
}

type WorkerConfigurationSchema struct {
	Version  int `json:"version"`
	Services struct {
		RabbitMQ RabbitMQConfiguration `yaml:"rabbit_mq"`
	} `yaml:"services"`
	Workdir string `yaml:"workdir"`
	Void    bool   `yaml:"void"`
	Docker  struct {
		MaxMemory bytefmt.ByteSize `yaml:"max_memory"`
		Timeout   time.Duration    `yaml:"timeout"`
	} `yaml:"docker"`
}

type ConfigurationSchema struct {
	Server ServerConfigurationSchema `yaml:"server"`
	Worker WorkerConfigurationSchema `yaml:"worker"`
}

var Configuration *ConfigurationSchema

func ParseConfiguration(filename string) (*ConfigurationSchema, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &ConfigurationSchema{}
	err = yaml.Unmarshal(yamlFile, cfg)
	return cfg, err
}

func MustFindAndReadConfiguration() {
	config_filename := os.Getenv("INFOMARK_CONFIG_FILE")

	if config_filename == "" {
		log.Fatalf("Env-var INFOMARK_CONFIG_FILE not given")
	}
	var err error

	Configuration, err = ParseConfiguration(config_filename)

	if err != nil {
		panic(err)
	}

}
