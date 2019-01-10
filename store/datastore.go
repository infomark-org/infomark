// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

package store

import (
	"fmt"
	"log"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	DriverSqlite = "sqlite3"
)

type Datastore interface {
	GetUserFromIdString(userID string) (*model.User, error)

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

	db, err := gorm.Open(DriverSqlite, "test.db")
	if err != nil {
		log.Fatal("Failed to init db:", err)
	}

	db.LogMode(true)

	if err := db.AutoMigrate(&model.User{}).Error; err != nil {
		txt := "AutoMigrate Users table failed"
		panic(fmt.Sprintf("%s: %s", txt, err))
	}

	return &datastore{db: db}
}
