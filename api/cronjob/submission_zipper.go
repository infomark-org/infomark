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
  "fmt"
  "strconv"

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

func FetchStudentSubmissions(groupID int64, taskID int64) ([]StudentSubmission, error) {
  p := []StudentSubmission{}
  err := s.db.Select(&p, `
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
    fmt.Println(sheet.DueAt)
    if app.OverTime(sheet.DueAt) {
      fmt.Println("work on ", sheet.ID)
      sheet_lock_path := fmt.Sprintf("%s/infomark-sheet%s.lock", job.Directory, strconv.FormatInt(sheet.ID, 10))
      fmt.Println("create file %s", sheet_lock_path)
      if !helper.FileExists(sheet_lock_path) {
        fmt.Println(sheet.DueAt)
        helper.FileTouch(sheet_lock_path)

        courseID := int64(0)
        job.DB.Get(&courseID, "SELECT course_id FROM sheet_course WHERE sheet_id = $1;", sheet.ID)

        fmt.Println("course ", courseID)

        groups, _ := job.Stores.Group.GroupsOfCourse(courseID)
        tasks, _ := job.Stores.Task.TasksOfSheet(sheet.ID)

        for _, task := range tasks {
          for _, group := range groups {
            dest_zip_lock := fmt.Sprintf("%s/infomark-sheet%s-task%s-group%s.lock",
              job.Directory,
              strconv.FormatInt(sheet.ID, 10),
              strconv.FormatInt(task.ID, 10),
              strconv.FormatInt(group.ID, 10),
            )

            if !helper.FileExists(dest_zip_lock) {
              // we gonna zip all submissions from students in group x for task y
              helper.FileTouch(dest_zip_lock)
            }

            submissions, _ := FetchStudentSubmissions(group.ID, task.ID)

          }

        }

      } else {
        fmt.Println("already done", sheet.ID)
      }
    } else {
      fmt.Println("ok", sheet.ID)
    }

  }
}
