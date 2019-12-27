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

package configuration

import (
	"github.com/infomark-org/infomark-backend/configuration/bytefmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type RabbitMQConfiguration struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Key      string `yaml:"key"`
}

type ServerConfiguration struct {
	Version   string `json:"version"`
	Debugging struct {
		Enabled    bool   `yaml:"enabled"`
		UserID     int64  `yaml:"user_id"`
		UserIsRoot bool   `yaml:"user_is_root"`
		LogLevel   string `yaml:"log_level"`
	} `yaml:"debugging"`
	HTTP struct {
		UseHTTPS bool   `yaml:"use_https"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Domain   string `yaml:"domain"`
		Timeouts struct {
			Read  time.Duration `yaml:"read"`
			Write time.Duration `yaml:"write"`
		} `yaml:"timeouts"`
		Limits struct {
			MaxHeader      bytefmt.ByteSize `yaml:"max_header"`
			MaxRequestJSON bytefmt.ByteSize `yaml:"max_request_json"`
			MaxSubmission  bytefmt.ByteSize `yaml:"max_submission"`
		} `yaml:"limits"`
	} `yaml:"http"`
	DistributeJobs bool `yaml:"distribute_jobs"`
	Authentication struct {
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
		TotalRequestsPerMinute int `yaml:"total_requests_per_minute"`
	} `yaml:"authentication"`
	Cronjobs struct {
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
	Paths struct {
		Uploads        string `yaml:"uploads"`
		Common         string `yaml:"common"`
		GeneratedFiles string `yaml:"generated_files"`
		Fixtures       string `yaml:"fixtures"`
	} `yaml:"paths"`
}

type WorkerConfiguration struct {
	Version  string `json:"version"`
	Services struct {
		RabbitMQ RabbitMQConfiguration `yaml:"rabbit_mq"`
	} `yaml:"services"`
	Workdir string `yaml:"workdir"`
	Void    bool   `yaml:"void"`
	Docker  struct {
		MaxMemory bytefmt.ByteSize `yaml:"max_memory"`
	} `yaml:"docker"`
}

type Configuration struct {
	Server ServerConfiguration `yaml:"server"`
	Worker WorkerConfiguration `yaml:"worker"`
}

func ParseConfiguration(filename string) (*Configuration, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &Configuration{}
	err = yaml.Unmarshal(yamlFile, cfg)
	return cfg, err
}
