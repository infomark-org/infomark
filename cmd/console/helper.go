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

package console

import (
	"log"
	"strconv"

	"github.com/infomark-org/infomark-backend/api/app"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

func failWhenSmallestWhiff(err error) {
	if err != nil {
		panic(err)
	}
}

func ConnectAndStores() (*sqlx.DB, *app.Stores, error) {

	db, err := sqlx.Connect("postgres", viper.GetString("database_connection"))
	if err != nil {
		return nil, nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, nil, err
	}

	stores := app.NewStores(db)
	return db, stores, nil
}

func MustConnectAndStores() (*sqlx.DB, *app.Stores) {

	db, stores, err := ConnectAndStores()
	failWhenSmallestWhiff(err)
	return db, stores
}

func MustInt64Parameter(argStr string, name string) int64 {
	argInt, err := strconv.Atoi(argStr)
	if err != nil {
		log.Fatalf("cannot convert %s '%s' to int64\n", name, argStr)
		return int64(0)
	}
	return int64(argInt)
}

func MustIntParameter(argStr string, name string) int {
	argInt, err := strconv.Atoi(argStr)
	if err != nil {
		log.Fatalf("cannot convert %s '%s' to int\n", name, argStr)
		return int(0)
	}
	return int(argInt)
}
