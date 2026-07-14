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

// This is heavily inspired by meddler.
// We basically rely on sqlx but use this library to build the missing statements

package migration

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// migrationFiles embeds the SQL migration scripts stored in the data
// subdirectory directly into the binary. This replaces the previous
// markbates/pkger loader, which resolved the module root at runtime by
// shelling out to the go tool and therefore failed with "open
// /migration/data: no such file or directory" whenever the go binary was not
// installed on the host (see SESSION.md). Embedding makes the migration data
// part of the compiled artifact, so migrations run without any external
// tooling.
//
//go:embed data/*.sql
var migrationFiles embed.FS

func UpdateDatabase(db *sqlx.DB, log *logrus.Logger) {
	log.Info("configure databse migration...")

	// The iofs source driver reads the migration scripts from the embedded
	// filesystem rooted at the data directory, mirroring the layout pkger
	// previously served from "/migration/data".
	fileDriver, err := iofs.New(migrationFiles, "data")
	if err != nil {
		log.Fatal(err)
	}

	databaseDriver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatal(err)

	}

	m, err := migrate.NewWithInstance("iofs", fileDriver, "postgres", databaseDriver)
	if err != nil {
		log.Fatal(err)

	}

	database_schema_version, dirty, err := m.Version()
	if err != migrate.ErrNilVersion && err != nil {
		log.Fatal(err)
	}

	log.WithFields(logrus.Fields{
		"database_schema_version": database_schema_version,
		"dirty":                   dirty,
	}).Info("before migration")

	err = m.Up()
	if err != migrate.ErrNoChange && err != nil {
		log.Fatal(err)
	}

	database_schema_version, dirty, err = m.Version()
	if err != migrate.ErrNilVersion && err != nil {
		log.Fatal(err)
	}

	log.WithFields(logrus.Fields{
		"database_schema_version": database_schema_version,
		"dirty":                   dirty,
	}).Info("after migration")

}
