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
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/infomark-org/infomark/api/helper"
	"github.com/infomark-org/infomark/configuration"
	"github.com/infomark-org/infomark/migration"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	DatabaseCmd.AddCommand(DatabaseRunCmd)
	DatabaseCmd.AddCommand(DatabaseRestoreCmd)
	DatabaseCmd.AddCommand(DatabaseBackupCmd)
	DatabaseCmd.AddCommand(DatabaseMigrateCmd)
}

var DatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Management of database.",
}

var DatabaseRunCmd = &cobra.Command{
	Use:   "run [sql]",
	Short: "run a sql command",
	Long:  `run a SQl statement. This statement will persistently changes entries!`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sql := args[0]

		configuration.MustFindAndReadConfiguration()

		// PGPASSWORD=pass psql -hlocalhost -Uuser -p5433 -d db -c "..."
		shell := exec.Command("psql",
			"-h", configuration.Configuration.Server.Services.Postgres.Host,
			"-U", configuration.Configuration.Server.Services.Postgres.User,
			"-p", fmt.Sprintf("%v", configuration.Configuration.Server.Services.Postgres.Port),
			"-d", configuration.Configuration.Server.Services.Postgres.Database,
			"-c", sql)
		shell.Env = os.Environ()
		shell.Env = append(shell.Env, fmt.Sprintf("PGPASSWORD=%s", configuration.Configuration.Server.Services.Postgres.Password))
		out, err := shell.CombinedOutput()
		fmt.Printf("%s", out)
		if err != nil {
			log.Fatal("executing SQL-statement was not successful")
		}

	},
}

var DatabaseRestoreCmd = &cobra.Command{
	Use:   "restore [file.sql.gz]",
	Short: "restore database from a file",
	Long:  `Will clean entire database and load a snapshot`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]

		if !helper.FileExists(file) {
			log.Fatalf("The file %s does not exists!\n", file)
		}

		configuration.MustFindAndReadConfiguration()

		// // dbname := args[1]

		// dropdb -h ${POSTGRES_HOST} -p ${POSTGRES_PORT} -U ${POSTGRES_USER} db
		// PGPASSWORD=pass dropdb -hlocalhost -Uuser -p5433 db
		shell := exec.Command("dropdb",
			"-h", configuration.Configuration.Server.Services.Postgres.Host,
			"-U", configuration.Configuration.Server.Services.Postgres.User,
			"-p", fmt.Sprintf("%v", configuration.Configuration.Server.Services.Postgres.Port),
			configuration.Configuration.Server.Services.Postgres.Database)
		shell.Env = os.Environ()
		shell.Env = append(shell.Env, fmt.Sprintf("PGPASSWORD=%s", configuration.Configuration.Server.Services.Postgres.Password))
		out, err := shell.CombinedOutput()
		fmt.Printf("%s", out)
		if err != nil {
			log.Fatal("dropping db was not successful")
		}

		// createdb  -h ${POSTGRES_HOST} -p ${POSTGRES_PORT} -U ${POSTGRES_USER} --owner="${POSTGRES_USER}" ${POSTGRES_DB}
		// PGPASSWORD=pass createdb -hlocalhost -Uuser -p5433 --owner=user db
		shell = exec.Command("createdb",
			"-h", configuration.Configuration.Server.Services.Postgres.Host,
			"-U", configuration.Configuration.Server.Services.Postgres.User,
			"-p", fmt.Sprintf("%v", configuration.Configuration.Server.Services.Postgres.Port),
			"--owner", configuration.Configuration.Server.Services.Postgres.User,
			configuration.Configuration.Server.Services.Postgres.Database)
		shell.Env = os.Environ()
		shell.Env = append(shell.Env, fmt.Sprintf("PGPASSWORD=%s", configuration.Configuration.Server.Services.Postgres.Password))
		out, err = shell.CombinedOutput()
		fmt.Printf("%s", out)
		if err != nil {
			log.Fatal("creating db was not successful")
		}

		// gunzip -c "${backup_filename}" | psql -h ${POSTGRES_HOST} -p ${POSTGRES_PORT} -U ${POSTGRES_USER} "${POSTGRES_DB}"
		shell1 := exec.Command("gunzip",
			"-c", file)
		shell2 := exec.Command("psql",
			"-h", configuration.Configuration.Server.Services.Postgres.Host,
			"-U", configuration.Configuration.Server.Services.Postgres.User,
			"-p", fmt.Sprintf("%v", configuration.Configuration.Server.Services.Postgres.Port),
			configuration.Configuration.Server.Services.Postgres.Database)
		shell2.Env = os.Environ()
		shell2.Env = append(shell2.Env, fmt.Sprintf("PGPASSWORD=%s", configuration.Configuration.Server.Services.Postgres.Password))

		r, w := io.Pipe()
		shell1.Stdout = w
		shell2.Stdin = r

		var b2 bytes.Buffer
		shell2.Stdout = &b2

		shell1.Start()
		shell2.Start()
		shell1.Wait()
		w.Close()
		err = shell2.Wait()
		// io.Copy(os.Stdout, &b2)
		if err != nil {
			log.Fatalf("load db from was not successful\n %s", err)
		}
	},
}

var DatabaseBackupCmd = &cobra.Command{
	Use:   "backup [file.sql.gz]",
	Short: "backup database to a file",
	Long:  `Will dump the entire database to a snapshot file`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file := args[0]

		if helper.FileExists(file) {
			log.Fatalf("The file %s does exists! Will not override!\n", file)
		}

		configuration.MustFindAndReadConfiguration()

		// export backup_filename="infomark_$(date +'%Y_%m_%dT%H_%M_%S').sql.gz"
		// pg_dump -h ${POSTGRES_HOST} -U ${POSTGRES_USER} -d ${POSTGRES_DB} -p ${POSTGRES_PORT} | gzip > ${backup_filename}
		// PGPASSWORD=pass pg_dump -h localhost -U user -d db -p 5433 | gzip > ${backup_filename}
		// shell1 := exec.Command("echo", "hi")
		shell1 := exec.Command("pg_dump",
			"-h", configuration.Configuration.Server.Services.Postgres.Host,
			"-U", configuration.Configuration.Server.Services.Postgres.User,
			"-p", fmt.Sprintf("%v", configuration.Configuration.Server.Services.Postgres.Port),
			"-d", configuration.Configuration.Server.Services.Postgres.Database)
		shell1.Env = os.Environ()
		shell1.Env = append(shell1.Env, fmt.Sprintf("PGPASSWORD=%s", configuration.Configuration.Server.Services.Postgres.Password))

		shell2 := exec.Command("gzip")

		r, w := io.Pipe()
		shell1.Stdout = w
		shell2.Stdin = r

		var b2 bytes.Buffer
		shell2.Stdout = &b2

		err := shell1.Start()
		if err != nil {
			log.Fatalf("Cannot call pg_dump and gzip! Did your run 'sudo apt install postgresql-client gzip'\n")
		}
		shell2.Start()
		err = shell1.Wait()
		if err != nil {
			log.Fatalf("Cannot call pg_dump and gzip! Did your run 'sudo apt install postgresql-client gzip'\n")
		}
		w.Close()
		err = shell2.Wait()
		if err != nil {
			panic(err)
		}

		destination, err := os.Create(file)
		if err != nil {
			panic(err)
		}
		defer destination.Close()
		_, err = io.Copy(destination, &b2)
		if err != nil {
			log.Fatalf("storing snapshot was not successful\n %s", err)
		}

	},
}

var DatabaseMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate database to latest version",
	Long:  `manually run database migration`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		log := logrus.New()
		log.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
		log.Out = os.Stdout

		configuration.MustFindAndReadConfiguration()

		db, err := sqlx.Connect("postgres", configuration.Configuration.Server.PostgresURL())
		if err != nil {
			log.WithField("module", "database").Error(err)
		}

		migration.UpdateDatabase(db, log)

	},
}
