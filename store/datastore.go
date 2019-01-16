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

package store

import (
	"fmt"
	"log"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Datastore interface {
	// retrieve user model by id from url parameter
	CountUsers() (int, error)
	GetUserFromIdString(userID string) (*model.User, error)
	GetUserFromEmail(email string) (*model.User, error)

	ORM() *gorm.DB
}

func (ds *datastore) ORM() *gorm.DB {
	return ds.db
}

type datastore struct {
	db *gorm.DB
}

var ds = newDatastore()

func DS() Datastore { return ds }
func ORM() *gorm.DB { return ds.ORM() }

func newDatastore() Datastore {

	db, err := gorm.Open("postgres", "host=localhost user=infomark_user dbname=infomark_db password=infomark_password")
	if err != nil {
		log.Fatal("Failed to init db:", err)
	}

	// db.LogMode(true)

	if err := db.AutoMigrate(&model.User{}).Error; err != nil {
		txt := "AutoMigrate Users table failed"
		panic(fmt.Sprintf("%s: %s", txt, err))
	}

	return &datastore{db: db}
}
