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

package helper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"strings"

	txdb "github.com/DATA-DOG/go-txdb"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func StringArrayToIntArray(values []string) ([]int, error) {
	out := make([]int, len(values))
	for index, value := range values {
		v, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		out[index] = v
	}
	return out, nil
}

func StringArrayFromUrl(r *http.Request, name string, standard []string) []string {
	rolesFromURL, ok := r.URL.Query()[name]
	if ok {
		return strings.Split(rolesFromURL[0], ",")
	} else {
		return standard
	}
}

func StringFromUrl(r *http.Request, name string, standard string) string {
	rolesFromURL, ok := r.URL.Query()[name]
	if ok {
		if len(rolesFromURL[0]) > 0 {
			return rolesFromURL[0]
		} else {
			return standard
		}
	} else {
		return standard
	}
}

func IntFromUrl(r *http.Request, name string, standard int) int {
	str := StringFromUrl(r, name, "uglyhardcoded")
	if str == "uglyhardcoded" {
		return standard
	} else {
		i, err := strconv.Atoi(str)
		if err != nil {
			return standard
		} else {
			return i
		}
	}
}

func Int64FromUrl(r *http.Request, name string, standard int64) int64 {
	str := StringFromUrl(r, name, "uglyhardcoded")
	if str == "uglyhardcoded" {
		return standard
	} else {
		i, err := strconv.Atoi(str)
		if err != nil {
			return standard
		} else {
			return int64(i)
		}
	}
}

// similar to gin.H as a neat wrapper
type H map[string]interface{}

// for testing convert any model to SimulateRequest
func ToH(z interface{}) map[string]interface{} {
	data, _ := json.Marshal(z)
	var msgMapTemplate interface{}
	_ = json.Unmarshal(data, &msgMapTemplate)
	return msgMapTemplate.(map[string]interface{})
}

// Time return time.Now() but without nanseconds for passing unit-tests
func Time(t time.Time) time.Time {
	format := "2006-01-02 15:04:05 +0000 CET"
	R, _ := time.Parse(format, t.Format(format))
	return R
}

// var tokenManager *authenticate.TokenAuth

func SetConfigFile() {

	// Find home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Search config in home directory with name ".go-base" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigName(".infomark-backend")
}
func InitConfig() {

	SetConfigFile()
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// func init() {
// 	tokenManager, _ = authenticate.NewTokenAuth()
// }

func init() {
	InitConfig()
	// we register an sql driver named "txdb"
	// This allows to run all tests as transaction in isolated environemnts to make sure
	// we do not accidentially alter the database in a persistent way. Hence,  all tests can run
	// in an arbitrary order.

	txdb.Register("psql_txdb", "postgres", viper.GetString("database_connection"))
}

// TransactionDB creates a sql-driver which seemlessly supports transactions.
func TransactionDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("psql_txdb", "identifier")

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, err
}
