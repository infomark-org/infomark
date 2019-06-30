// InfoMark - a platform for managing courses with
//            distributing exercise materials and testing exercise submissions
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

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/franela/goblin"
	"github.com/spf13/viper"
)

func TestSubmission(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail
	// DefaultSubmissionProducer = &VoidProducer{}

	tape := &Tape{}

	var stores *Stores

	studentJWT := NewJWTRequest(112, false)
	otherStudentJWT := NewJWTRequest(113, false)
	tutorJWT := NewJWTRequest(2, false)
	adminJWT := NewJWTRequest(1, true)

	g.Describe("Submission", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
			_ = stores
		})

		g.It("Query should require access claims", func() {

			w := tape.Get("/api/v1/courses/1/tasks/1/submission")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)
		})

		g.It("Tutors can download a collection of submissions", func() {

			courseID := int64(1)
			taskID := int64(1)
			groupID := int64(1)

			sheet, err := stores.Task.IdentifySheetOfTask(taskID)
			g.Assert(err).Equal(nil)

			hnd := helper.NewSubmissionsCollectionFileHandle(courseID, sheet.ID, taskID, groupID)

			defer hnd.Delete()

			// no files so far
			g.Assert(hnd.Exists()).Equal(false)

			src := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			copyFile(src, hnd.Path())

			g.Assert(hnd.Exists()).Equal(true)

			w := tape.Get("/api/v1/courses/1/tasks/1/groups/1/file")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/tasks/1/groups/1/file", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Get("/api/v1/courses/1/tasks/1/groups/1/file", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			w = tape.Get("/api/v1/courses/1/tasks/1/groups/1/file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Students can upload solution (create)", func() {

			deadlineAt := NowUTC().Add(time.Hour)
			publishedAt := NowUTC().Add(-time.Hour)

			// make sure the upload date is good
			task, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)
			sheet, err := stores.Task.IdentifySheetOfTask(task.ID)
			g.Assert(err).Equal(nil)

			sheet.PublishAt = publishedAt
			sheet.DueAt = deadlineAt
			err = stores.Sheet.Update(sheet)
			g.Assert(err).Equal(nil)

			defer helper.NewSubmissionFileHandle(3001).Delete()

			// no files so far
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

			// remove all submission from student
			_, err = tape.DB.Exec("DELETE FROM submissions WHERE user_id = 112;")
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)

			// upload
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			w, err = tape.Upload("/api/v1/courses/1/tasks/1/submission", filename, "application/zip", studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			createdSubmission, err := stores.Submission.GetByUserAndTask(112, 1)
			g.Assert(err).Equal(nil)

			g.Assert(helper.NewSubmissionFileHandle(createdSubmission.ID).Exists()).Equal(true)
			defer helper.NewSubmissionFileHandle(createdSubmission.ID).Delete()

			// files exists
			w = tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Students cannot upload solution (create) since too late", func() {

			deadlineAt := NowUTC().Add(-2 * time.Hour)
			publishedAt := NowUTC().Add(-10 * time.Hour)

			// make sure the upload date is good
			task, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)
			sheet, err := stores.Task.IdentifySheetOfTask(task.ID)
			g.Assert(err).Equal(nil)

			sheet.PublishAt = publishedAt
			sheet.DueAt = deadlineAt
			err = stores.Sheet.Update(sheet)
			g.Assert(err).Equal(nil)

			defer helper.NewSubmissionFileHandle(3001).Delete()

			// no files so far
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

			// remove all submission from student
			_, err = tape.DB.Exec("DELETE FROM submissions WHERE user_id = 112;")
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)

			// upload
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			w, err = tape.Upload("/api/v1/courses/1/tasks/1/submission", filename, "application/zip", studentJWT)

			g.Assert(err).Equal(nil)

			g.Assert(w.Code).Equal(http.StatusBadRequest)

			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)
			defer helper.NewSubmissionFileHandle(3001).Delete()

		})

		g.It("Students can upload solution (update)", func() {

			defer helper.NewSubmissionFileHandle(3001).Delete()

			deadlineAt := NowUTC().Add(time.Hour)
			publishedAt := NowUTC().Add(-time.Hour)

			// make sure the upload date is good
			task, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)
			sheet, err := stores.Task.IdentifySheetOfTask(task.ID)
			g.Assert(err).Equal(nil)

			sheet.PublishAt = publishedAt
			sheet.DueAt = deadlineAt
			err = stores.Sheet.Update(sheet)
			g.Assert(err).Equal(nil)

			// no files so far
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

			// upload
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			w, err := tape.Upload("/api/v1/courses/1/tasks/1/submission", filename, "application/zip", studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(true)

			// files exists
			w = tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Students cannot upload solution (update) too late", func() {

			defer helper.NewSubmissionFileHandle(3001).Delete()

			deadlineAt := NowUTC().Add(-time.Hour)
			publishedAt := NowUTC().Add(-2 * time.Hour)

			// make sure the upload date is good
			task, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)
			sheet, err := stores.Task.IdentifySheetOfTask(task.ID)
			g.Assert(err).Equal(nil)

			sheet.PublishAt = publishedAt
			sheet.DueAt = deadlineAt
			err = stores.Sheet.Update(sheet)
			g.Assert(err).Equal(nil)

			// no files so far
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

			// upload
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			w, err := tape.Upload("/api/v1/courses/1/tasks/1/submission", filename, "application/zip", studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

		})

		g.It("creating a submission will crate an empty grade entry as well", func() {

			defer helper.NewSubmissionFileHandle(3001).Delete()

			// no files so far
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

			// remove all submission from student
			_, err := tape.DB.Exec("DELETE FROM submissions WHERE user_id = 112;")
			g.Assert(err).Equal(nil)

			// remove all grades from student
			_, err = tape.DB.Exec("TRUNCATE TABLE grades;")
			g.Assert(err).Equal(nil)

			// no submission
			w := tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)

			// upload
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			w, err = tape.Upload("/api/v1/courses/1/tasks/1/submission", filename, "application/zip", studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			createdSubmission, err := stores.Submission.GetByUserAndTask(112, 1)
			g.Assert(err).Equal(nil)

			g.Assert(helper.NewSubmissionFileHandle(createdSubmission.ID).Exists()).Equal(true)
			defer helper.NewSubmissionFileHandle(createdSubmission.ID).Delete()

			// files exists
			w = tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// verify there is also a grade
			state := int64(88)
			err = tape.DB.Get(&state, "SELECT public_execution_state from grades WHERE submission_id = $1;", createdSubmission.ID)
			g.Assert(err).Equal(nil)
			g.Assert(state).Equal(int64(0))
		})

		g.It("Students can only access their own submissions", func() {

			defer helper.NewSubmissionFileHandle(3001).Delete()

			// no files so far
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

			// upload
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			w, err := tape.Upload("/api/v1/courses/1/tasks/1/submission", filename, "application/zip", studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(true)

			// access own submission
			w = tape.Get("/api/v1/courses/1/submissions/3001/file", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// access others submission
			w = tape.Get("/api/v1/courses/1/submissions/3001/file", otherStudentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

		})

		g.It("Admins can upload solution for a student (even if it is too late)", func() {

			studentJWT := NewJWTRequest(112, false)
			adminJWT := NewJWTRequest(1, true)

			deadlineAt := NowUTC().Add(-time.Hour)
			publishedAt := NowUTC().Add(-2 * time.Hour)

			// make sure the upload date is good
			task, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)
			sheet, err := stores.Task.IdentifySheetOfTask(task.ID)
			g.Assert(err).Equal(nil)

			sheet.PublishAt = publishedAt
			sheet.DueAt = deadlineAt
			err = stores.Sheet.Update(sheet)
			g.Assert(err).Equal(nil)

			defer helper.NewSubmissionFileHandle(3001).Delete()

			// no files so far
			g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

			// remove all submission from student
			_, err = tape.DB.Exec("DELETE FROM submissions WHERE user_id = 112;")
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)

			// upload
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			params := map[string]string{
				"user_id": "112",
			}

			// upload as admin
			w, err = tape.UploadWithParameters("/api/v1/courses/1/tasks/1/submission", filename, "application/zip", params, adminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			createdSubmission, err := stores.Submission.GetByUserAndTask(112, 1)
			g.Assert(err).Equal(nil)
			g.Assert(createdSubmission.UserID).Equal(int64(112))

			g.Assert(helper.NewSubmissionFileHandle(createdSubmission.ID).Exists()).Equal(true)
			defer helper.NewSubmissionFileHandle(createdSubmission.ID).Delete()

			// files exists
			w = tape.Get("/api/v1/courses/1/tasks/1/submission", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("tutors/admins can filter submissions", func() {

			w := tape.Get("/api/v1/courses/1/submissions")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/submissions", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Get("/api/v1/courses/1/submissions", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			submissionsAllActual := []SubmissionResponse{}
			err := json.NewDecoder(w.Body).Decode(&submissionsAllActual)
			g.Assert(err).Equal(nil)

			w = tape.Get("/api/v1/courses/1/submissions?group_id=4", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			submissionsG4Actual := []SubmissionResponse{}
			err = json.NewDecoder(w.Body).Decode(&submissionsG4Actual)
			g.Assert(err).Equal(nil)

			w = tape.Get("/api/v1/courses/1/submissions?task_id=2", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			submissionsT4Actual := []SubmissionResponse{}
			err = json.NewDecoder(w.Body).Decode(&submissionsT4Actual)
			g.Assert(err).Equal(nil)

			for _, el := range submissionsT4Actual {
				g.Assert(el.TaskID).Equal(int64(2))
			}

			w = tape.Get("/api/v1/courses/1/submissions?user_id=112", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			submissionsU112Actual := []SubmissionResponse{}
			err = json.NewDecoder(w.Body).Decode(&submissionsU112Actual)
			g.Assert(err).Equal(nil)

			for _, el := range submissionsU112Actual {
				g.Assert(el.UserID).Equal(int64(112))
			}

		})

		g.AfterEach(func() {
			tape.AfterEach()
		})
	})

}
