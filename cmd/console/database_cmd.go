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
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func parseConnectionString(conn string) map[string]string {
	if len(conn) == 0 {
		log.Fatal("database_connection-string from config is empty")
	}
	connParts, err := pq.ParseURL(conn)
	failWhenSmallestWhiff(err)
	parts := strings.Split(connParts, " ")

	infoMap := make(map[string]string)
	for _, part := range parts {
		v := strings.Split(part, "=")
		if len(v) >= 2 {
			switch v[0] {
			case "dbname", "host", "password", "port", "user":
				infoMap[v[0]] = v[1]
			}
		}

	}
	return infoMap
}

func init() {
	DatabaseCmd.AddCommand(DatabaseRunCmd)
	DatabaseCmd.AddCommand(DatabaseRestoreCmd)
	DatabaseCmd.AddCommand(DatabaseBackupCmd)
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

		infoMap := parseConnectionString(viper.GetString("database_connection"))

		// PGPASSWORD=pass psql -hlocalhost -Uuser -p5433 -d db -c "..."
		shell := exec.Command("psql",
			"-h", infoMap["host"],
			"-U", infoMap["user"],
			"-p", infoMap["port"],
			"-d", infoMap["dbname"],
			"-c", sql)
		shell.Env = os.Environ()
		shell.Env = append(shell.Env, fmt.Sprintf("PGPASSWORD=%s", infoMap["password"]))
		out, err := shell.CombinedOutput()
		fmt.Printf("%s", out)
		if err != nil {
			log.Fatal("executing SQL-statement was not successfull")
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

		infoMap := parseConnectionString(viper.GetString("database_connection"))

		// // dbname := args[1]

		// dropdb -h ${POSTGRES_HOST} -p ${POSTGRES_PORT} -U ${POSTGRES_USER} db
		// PGPASSWORD=pass dropdb -hlocalhost -Uuser -p5433 db
		shell := exec.Command("dropdb",
			"-h", infoMap["host"],
			"-U", infoMap["user"],
			"-p", infoMap["port"],
			infoMap["dbname"])
		shell.Env = os.Environ()
		shell.Env = append(shell.Env, fmt.Sprintf("PGPASSWORD=%s", infoMap["password"]))
		out, err := shell.CombinedOutput()
		fmt.Printf("%s", out)
		if err != nil {
			log.Fatal("dropping db was not successfull")
		}

		// createdb  -h ${POSTGRES_HOST} -p ${POSTGRES_PORT} -U ${POSTGRES_USER} --owner="${POSTGRES_USER}" ${POSTGRES_DB}
		// PGPASSWORD=pass createdb -hlocalhost -Uuser -p5433 --owner=user db
		shell = exec.Command("createdb",
			"-h", infoMap["host"],
			"-U", infoMap["user"],
			"-p", infoMap["port"],
			"--owner", infoMap["user"],
			infoMap["dbname"])
		shell.Env = os.Environ()
		shell.Env = append(shell.Env, fmt.Sprintf("PGPASSWORD=%s", infoMap["password"]))
		out, err = shell.CombinedOutput()
		fmt.Printf("%s", out)
		if err != nil {
			log.Fatal("creating db was not successfull")
		}

		// gunzip -c "${backup_filename}" | psql -h ${POSTGRES_HOST} -p ${POSTGRES_PORT} -U ${POSTGRES_USER} "${POSTGRES_DB}"
		shell1 := exec.Command("gunzip",
			"-c", file)
		shell2 := exec.Command("psql",
			"-h", infoMap["host"],
			"-U", infoMap["user"],
			"-p", infoMap["port"],
			infoMap["dbname"])
		shell2.Env = os.Environ()
		shell2.Env = append(shell2.Env, fmt.Sprintf("PGPASSWORD=%s", infoMap["password"]))

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
			log.Fatal("load db from was not successfull\n %s", err)
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

		infoMap := parseConnectionString(viper.GetString("database_connection"))

		// export backup_filename="infomark_$(date +'%Y_%m_%dT%H_%M_%S').sql.gz"
		// pg_dump -h ${POSTGRES_HOST} -U ${POSTGRES_USER} -d ${POSTGRES_DB} -p ${POSTGRES_PORT} | gzip > ${backup_filename}
		// PGPASSWORD=pass pg_dump -h localhost -U user -d db -p 5433 | gzip > ${backup_filename}
		// shell1 := exec.Command("echo", "hi")
		shell1 := exec.Command("pg_dump",
			"-h", infoMap["host"],
			"-U", infoMap["user"],
			"-p", infoMap["port"],
			"-d", infoMap["dbname"])
		shell1.Env = os.Environ()
		shell1.Env = append(shell1.Env, fmt.Sprintf("PGPASSWORD=%s", infoMap["password"]))

		shell2 := exec.Command("gzip")

		r, w := io.Pipe()
		shell1.Stdout = w
		shell2.Stdin = r

		var b2 bytes.Buffer
		shell2.Stdout = &b2

		shell1.Start()
		shell2.Start()
		shell1.Wait()
		w.Close()
		err := shell2.Wait()

		destination, err := os.Create(file)
		if err != nil {
			panic(err)
		}
		defer destination.Close()
		_, err = io.Copy(destination, &b2)
		if err != nil {
			log.Fatal("storing snapshot was not successfull\n %s", err)
		}

	},
}
