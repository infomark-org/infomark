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

package cronjob

import (
	"archive/zip"
	"fmt"
	"io"
	"os"

	"github.com/infomark-org/infomark/api/app"
	"github.com/infomark-org/infomark/api/helper"
	"github.com/infomark-org/infomark/configuration"
	"github.com/infomark-org/infomark/email"
	"github.com/jmoiron/sqlx"
)

// SubmissionFileZipper links all ressource to zip submissions
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

// StudentSubmission is a view from the database used to identify which
// submissions should be included in the final zip file
type StudentSubmission struct {
	ID              int64  `db:"id"`
	StudentLastName string `db:"last_name"`
}

// FetchStudentSubmissions queries the database to gather all submissions for a given group and task
func FetchStudentSubmissions(db *sqlx.DB, groupID int64, taskID int64) ([]StudentSubmission, error) {
	p := []StudentSubmission{}
	// 	err := db.Select(&p, `
	// SELECT s.id, u.first_name, u.last_name FROM submissions s
	//   INNER JOIN user_group ug ON ug.user_id = s.user_id
	//   INNER JOIN users u ON u.id  = s.user_id
	//   WHERE  ug.group_id = $1
	//   AND s.task_id = $2`, groupID, taskID)

	err := db.Select(&p, `
SELECT s.id, string_agg(last_names, '-') AS last_name
FROM   submissions AS s, LATERAL
       (SELECT DISTINCT ON (g.id) u.last_name
        FROM  users AS u,
              user_group AS ug,
              user_course AS uc,
              grades AS g
        WHERE  u.id = g.user_id
        AND    s.team_id = uc.team_id
        AND    s.id = g.submission_id
        AND    s.team_id = g.team_id
        AND    u.id = uc.user_id
        AND    u.id = g.user_id
        AND    uc.team_confirmed
        AND    ug.group_id = $1) AS _(last_names)
WHERE  s.task_id = $2
GROUP BY s.id
	`, groupID, taskID)
	return p, err
}

// Run executes a job to zip all submissions for each group and task
func (job *SubmissionFileZipper) Run() {
	// for each sheet
	// if ended
	//   touch generated/infomark-sheet{sheetID}.lock
	//   for each group
	//     touch generated/infomark-sheet{sheetID}-task{taskID}-group{groupID}-submissions.lock
	//     create generated/infomark-sheet{sheetID}-task{taskID}-group{groupID}-submissions.zip
	// zip nested task
	fmt.Println("Start zipping all submissions to:", job.Directory)

	sheets, _ := job.Stores.Sheet.GetAll()

	for _, sheet := range sheets {
		if app.OverTime(sheet.DueAt) {
			// fmt.Println("work on ", sheet.ID)
			sheetLockPath := fmt.Sprintf("%s/infomark-sheet%d.lock", job.Directory, sheet.ID)

			fmt.Printf("test lock file '%s'\n", sheetLockPath)

			if !helper.FileExists(sheetLockPath) {
				fmt.Println(" --> create", sheet.ID)
				helper.FileTouch(sheetLockPath)

				courseID := int64(0)
				job.DB.Get(&courseID, "SELECT course_id FROM sheet_course WHERE sheet_id = $1;", sheet.ID)

				groups, _ := job.Stores.Group.GroupsOfCourse(courseID)
				tasks, _ := job.Stores.Task.TasksOfSheet(sheet.ID)

				for _, task := range tasks {
					fmt.Println("  work on task ", task.ID)

					for _, group := range groups {
						archivLockPath := fmt.Sprintf("%s/collection-course%d-sheet%d-task%d-group%d.lock", job.Directory, courseID, sheet.ID, task.ID, group.ID)
						// archiv_zip_path := fmt.Sprintf("%s/infomark-course%d-sheet%d-task%d-group%d.zip", job.Directory, courseID, sheet.ID, task.ID, group.ID)

						archivZip := helper.NewSubmissionsCollectionFileHandle(courseID, sheet.ID, task.ID, group.ID)

						if !helper.FileExists(archivLockPath) && !archivZip.Exists() {

							// we gonna zip all submissions from students in group x for task y
							helper.FileTouch(archivLockPath)

							submissions, _ := FetchStudentSubmissions(job.DB, group.ID, task.ID)

							newZipFile, err := os.Create(archivZip.Path())
							if err != nil {
								// creating zip went wrong... delete lock files => retry later
								helper.FileDelete(archivLockPath)
								helper.FileDelete(sheetLockPath)
								fmt.Println(" Creating zip file went wrong.")
								return
							}
							defer newZipFile.Close()

							zipWriter := zip.NewWriter(newZipFile)
							defer zipWriter.Close()

							for _, submission := range submissions {
								// fmt.Println(submission)

								submissionHnd := helper.NewSubmissionFileHandle(submission.ID)

								// student did upload a zip file
								if submissionHnd.Exists() {
									// see https://stackoverflow.com/a/53802396/7443104
									// fmt.Println("add sbmission ", submission.ID, " to ", archivZip)

									// refer to the zip file
									zipfile, err := os.Open(submissionHnd.Path())
									if err != nil {
										// Open submission file failed
										// retry later
										helper.FileDelete(archivLockPath)
										helper.FileDelete(sheetLockPath)
										fmt.Println(" Opening submission file failed for ", submissionHnd.Path())
										return
									}
									defer zipfile.Close()

									// Get the file information
									info, err := zipfile.Stat()
									if err != nil {
										// Extracting file information failed
										// retry later
										helper.FileDelete(archivLockPath)
										helper.FileDelete(sheetLockPath)
										fmt.Println(" Extracting file information failed for", submissionHnd.Path())
										return
									}

									header, err := zip.FileInfoHeader(info)
									if err != nil {
										// Reading file header failed
										// retry later
										helper.FileDelete(archivLockPath)
										helper.FileDelete(sheetLockPath)
										fmt.Println(" Reading file header failed for", submissionHnd.Path())
										return
									}

									// Using FileInfoHeader() above only uses the basename of the file. If we want
									// to preserve the folder structure we can overwrite this with the full path.
									header.Name = fmt.Sprintf("%s.zip", submission.StudentLastName)

									// Change to deflate to gain better compression
									// see http://golang.org/pkg/archive/zip/#pkg-constants
									header.Method = zip.Deflate

									writer, err := zipWriter.CreateHeader(header)
									if err != nil {
										// Creating header failed
										// retry later
										helper.FileDelete(archivLockPath)
										helper.FileDelete(sheetLockPath)
										fmt.Println(" Creating header failed for", submissionHnd.Path())
										return
									}
									if _, err = io.Copy(writer, zipfile); err != nil {
										// Copying data to zipfile failed
										// retry later
										helper.FileDelete(archivLockPath)
										helper.FileDelete(sheetLockPath)
										fmt.Println(" Copying data to zip file failed for", submissionHnd.Path())
										return
									}
								}

							}

							// notify the tutor
							tutor, _ := job.Stores.User.Get(group.TutorID)

							from := configuration.Configuration.Server.Email.From

							email.DefaultMail.Send(email.NewEmail(from, tutor.Email, "Submission-Zip", fmt.Sprintf(`Hi %s,
the deadline for the exercise sheet '%s' is over. We have collected all submissions in a single zip file.
Please log in to grade these solutions.
`, tutor.FullName(), sheet.Name)))
						}
					}

				}
			} else {
				fmt.Println(" --> already done", sheet.ID)
			}

		}

	}
}
