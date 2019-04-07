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

package cronjob

import (
  "archive/zip"
  "fmt"
  "io"
  "os"

  "github.com/cgtuebingen/infomark-backend/api/app"
  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/jmoiron/sqlx"
)

type SubmissionFileZipper struct {
  Stores    *app.Stores
  DB        *sqlx.DB
  Directory string
}

// SELECT s.*, u.* FROM submissions s
// INNER JOIN user_group ug ON ug.user_id = s.user_id
// INNER JOIN users u ON u.id  = s.user_id
// WHERE  ug.group_id = 1

// LIMIT 10

type StudentSubmission struct {
  ID               int64  `db:"id"`
  StudentFirstName string `db:"first_name"`
  StudentLastName  string `db:"last_name"`
}

func FetchStudentSubmissions(db *sqlx.DB, groupID int64, taskID int64) ([]StudentSubmission, error) {
  p := []StudentSubmission{}
  err := db.Select(&p, `
  SELECT s.id, u.first_name, u.last_name FROM submissions s
  INNER JOIN user_group ug ON ug.user_id = s.user_id
  INNER JOIN users u ON u.id  = s.user_id
  WHERE  ug.group_id = $1
  AND s.task_id = $2`, groupID, taskID)
  return p, err
}

func (job *SubmissionFileZipper) Run() {
  // for each sheet
  // if ended
  //   touch generated/infomark-sheet{sheetID}.lock
  //   for each group
  //     touch generated/infomark-sheet{sheetID}-task{taskID}-group{groupID}-submissions.lock
  //     create generated/infomark-sheet{sheetID}-task{taskID}-group{groupID}-submissions.zip
  // zip nested task
  sheets, _ := job.Stores.Sheet.GetAll()

  for _, sheet := range sheets {
    if app.OverTime(sheet.DueAt) {
      // fmt.Println("work on ", sheet.ID)
      sheet_lock_path := fmt.Sprintf("%s/infomark-sheet%d.lock", job.Directory, sheet.ID)

      fmt.Printf("test lock file '%s'\n", sheet_lock_path)

      if !helper.FileExists(sheet_lock_path) {
        fmt.Println(" --> create", sheet.ID)
        helper.FileTouch(sheet_lock_path)

        courseID := int64(0)
        job.DB.Get(&courseID, "SELECT course_id FROM sheet_course WHERE sheet_id = $1;", sheet.ID)

        groups, _ := job.Stores.Group.GroupsOfCourse(courseID)
        tasks, _ := job.Stores.Task.TasksOfSheet(sheet.ID)

        for _, task := range tasks {
          fmt.Println("  work on task ", task.ID)

          for _, group := range groups {
            archiv_lock_path := fmt.Sprintf("%s/infomark-sheet%d-task%d-group%d.lock", job.Directory, sheet.ID, task.ID, group.ID)
            archiv_zip_path := fmt.Sprintf("%s/infomark-sheet%d-task%d-group%d.zip", job.Directory, sheet.ID, task.ID, group.ID)

            if !helper.FileExists(archiv_lock_path) && !helper.FileExists(archiv_zip_path) {

              // we gonna zip all submissions from students in group x for task y
              helper.FileTouch(archiv_lock_path)

              submissions, _ := FetchStudentSubmissions(job.DB, group.ID, task.ID)

              newZipFile, err := os.Create(archiv_zip_path)
              if err != nil {
                return
              }
              defer newZipFile.Close()

              zipWriter := zip.NewWriter(newZipFile)
              defer zipWriter.Close()

              for _, submission := range submissions {
                // fmt.Println(submission)

                submission_hnd := helper.NewSubmissionFileHandle(submission.ID)

                // student did upload a zip file
                if submission_hnd.Exists() {
                  // see https://stackoverflow.com/a/53802396/7443104
                  // fmt.Println("add sbmission ", submission.ID, " to ", archiv_zip_path)

                  // refer to the zip file
                  zipfile, err := os.Open(submission_hnd.Path())
                  if err != nil {
                    return
                  }
                  defer zipfile.Close()

                  // Get the file information
                  info, err := zipfile.Stat()
                  if err != nil {
                    return
                  }

                  header, err := zip.FileInfoHeader(info)
                  if err != nil {
                    return
                  }

                  // Using FileInfoHeader() above only uses the basename of the file. If we want
                  // to preserve the folder structure we can overwrite this with the full path.
                  header.Name = fmt.Sprintf("%s-%s.zip", submission.StudentLastName, submission.StudentFirstName)

                  // Change to deflate to gain better compression
                  // see http://golang.org/pkg/archive/zip/#pkg-constants
                  header.Method = zip.Deflate

                  writer, err := zipWriter.CreateHeader(header)
                  if err != nil {
                    return
                  }
                  if _, err = io.Copy(writer, zipfile); err != nil {
                    return
                  }
                }

              }
            }
          }

        }
      } else {
        fmt.Println(" --> already done", sheet.ID)
      }

    } else {
      // fmt.Println("ok", sheet.ID)
    }

  }
}
