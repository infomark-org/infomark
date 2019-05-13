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

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "infomark",
	Short: "A CI based course framework",
	Long: `InfoMark is a a scalable, modern and open-source
online course management system supporting auto-testing/grading of
programming assignments and distributing exercise sheets.
The infomark-server is the REST api backend for the course distributing system.

Complete documentation is available at https://infomark.org/.
	`,
}

var cfgFile = ""

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(InitConfig)
	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is infomark.yaml)")
}

// SetConfigFile searchs for a config file named ".informark.yml"
// which is located in the home-directory if the flag "--config" is not present.
func SetConfigFile() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		var err error
		// Find home directory.
		home := os.Getenv("INFOMARK_CONFIG_DIR")

		if home == "" {
			home, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
		}

		// Search config in home directory with name ".go-base" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".infomark")
	}

}

// initConfig reads in config file and ENV variables if set.
func InitConfig() {

	SetConfigFile()
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
