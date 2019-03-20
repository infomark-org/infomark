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

package app

import (
  "encoding/json"
  "fmt"
  "net/http"
  "testing"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/franela/goblin"
)

func TestGroup(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Group", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require access claims", func() {

      w := tape.Get("/api/v1/courses/1/groups")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/groups", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list all groups from a course", func() {
      w := tape.GetWithClaims("/api/v1/courses/1/groups", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      groups_actual := []GroupResponse{}
      err := json.NewDecoder(w.Body).Decode(&groups_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(groups_actual)).Equal(10)
    })

    g.It("Should get a specific group", func() {
      entry_expected, err := stores.Group.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/groups/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_actual := &GroupResponse{}
      err = json.NewDecoder(w.Body).Decode(entry_actual)
      g.Assert(err).Equal(nil)

      g.Assert(entry_actual.ID).Equal(entry_expected.ID)
      g.Assert(entry_actual.TutorID).Equal(entry_expected.TutorID)
      g.Assert(entry_actual.CourseID).Equal(entry_expected.CourseID)
      g.Assert(entry_actual.Description).Equal(entry_expected.Description)
    })

    g.It("Creating should require claims", func() {
      w := tape.Post("/api/v1/courses/1/groups", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.Xit("Creating should require body", func() {
      // TODO empty request with claims
    })

    g.It("Should create valid group", func() {
      entries_before, err := stores.Group.GroupsOfCourse(1)
      g.Assert(err).Equal(nil)

      entry_sent := &groupRequest{
        TutorID:     1,
        Description: "blah blahe",
      }

      err = entry_sent.Validate()
      g.Assert(err).Equal(nil)

      w := tape.PostWithClaims("/api/v1/courses/1/groups", helper.ToH(entry_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusCreated)

      entry_return := &GroupResponse{}
      err = json.NewDecoder(w.Body).Decode(&entry_return)
      g.Assert(entry_return.TutorID).Equal(entry_sent.TutorID)
      g.Assert(entry_return.CourseID).Equal(int64(1))
      g.Assert(entry_return.Description).Equal(entry_sent.Description)

      entries_after, err := stores.Group.GroupsOfCourse(1)
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) + 1)
    })

    g.It("Should update a group", func() {
      // group (id=1) belongs to course(id=1)
      entry_sent := &groupRequest{
        TutorID:     9,
        Description: "new descr",
      }

      // students
      w := tape.PlayDataWithClaims("PUT", "/api/v1/courses/1/groups/1", tape.ToH(entry_sent), 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PlayDataWithClaims("PUT", "/api/v1/courses/1/groups/1", tape.ToH(entry_sent), 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PlayDataWithClaims("PUT", "/api/v1/courses/1/groups/1", tape.ToH(entry_sent), 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Group.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.TutorID).Equal(entry_sent.TutorID)
      g.Assert(entry_after.CourseID).Equal(int64(1))
    })

    g.It("Should delete when valid access claims", func() {
      entries_before, err := stores.Group.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.Delete("/api/v1/courses/1/groups/1")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // verify nothing has changes
      entries_after, err := stores.Group.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before))

      // students
      w = tape.DeleteWithClaims("/api/v1/courses/1/groups/1", 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.DeleteWithClaims("/api/v1/courses/1/groups/1", 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.DeleteWithClaims("/api/v1/courses/1/groups/1", 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // verify a sheet less exists
      entries_after, err = stores.Group.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) - 1)
    })

    g.It("Should change a bid to a group", func() {
      userID := int64(112)

      // admins are not allowed
      w := tape.PostWithClaims("/api/v1/courses/1/groups/1/bids", H{"bid": 4}, 1, false)
      g.Assert(w.Code).Equal(http.StatusBadRequest)

      // tutors are not allowed
      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/bids", H{"bid": 4}, 2, false)
      g.Assert(w.Code).Equal(http.StatusBadRequest)

      // students
      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/bids", H{"bid": 4}, userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      // no content

      // delete to test insert
      tape.DB.Exec(`DELETE FROM group_bids where user_id = $1`, userID)

      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/bids", H{"bid": 4}, userID, false)
      g.Assert(w.Code).Equal(http.StatusCreated)
      entry_return := &GroupBidResponse{}
      err := json.NewDecoder(w.Body).Decode(&entry_return)
      g.Assert(err).Equal(nil)
      g.Assert(entry_return.Bid).Equal(4)

    })

    g.It("Find my group when being a student", func() {
      // a random student (checked via pgweb)
      loginID := int64(112)

      w := tape.Get("/api/v1/courses/1/group")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/group", loginID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_return := &GroupResponse{}
      err := json.NewDecoder(w.Body).Decode(&entry_return)
      g.Assert(err).Equal(nil)

      // we cannot check the other entries
      g.Assert(entry_return.CourseID).Equal(int64(1))
    })

    g.It("Find my group when being a tutor", func() {
      // a random student (checked via pgweb)
      loginID := int64(2)

      w := tape.Get("/api/v1/courses/1/group")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/group", loginID, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_return := &GroupResponse{}
      err := json.NewDecoder(w.Body).Decode(&entry_return)
      g.Assert(err).Equal(nil)

      // we cannot check the other entries
      g.Assert(entry_return.CourseID).Equal(int64(1))
      g.Assert(entry_return.TutorID).Equal(loginID)
    })

    g.It("Only tutors and admins can send emails to a group", func() {
      w := tape.Post("/api/v1/courses/1/groups/1/emails", H{"subject": "subj", "body": "body"})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // student
      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/emails", H{"subject": "subj", "body": "body"}, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutor
      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/emails", H{"subject": "subj", "body": "body"}, 2, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // admin
      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/emails", H{"subject": "subj", "body": "body"}, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

    })

    g.It("Permission test", func() {
      url := "/api/v1/courses/1/groups"

      // global root can do whatever they want
      w := tape.GetWithClaims(url, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      // enrolled tutors can access
      w = tape.GetWithClaims(url, 2, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // enrolled students can access
      w = tape.GetWithClaims(url, 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // disenroll student
      w = tape.DeleteWithClaims("/api/v1/courses/1/enrollments", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // cannot access anymore
      w = tape.GetWithClaims(url, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)
    })

    g.It("should manually add user to group (update)", func() {
      url := "/api/v1/courses/1/groups/1/enrollments"

      w := tape.Post(url, H{"user_id": 112})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // student
      w = tape.PostWithClaims(url, H{"user_id": 112}, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutor
      w = tape.PostWithClaims(url, H{"user_id": 112}, 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PostWithClaims(url, H{"user_id": 112}, 1, false)
      fmt.Println(w.Body)
      g.Assert(w.Code).Equal(http.StatusOK)

      enrollment, err := stores.Group.GetGroupEnrollmentOfUserInCourse(112, 1)
      g.Assert(err).Equal(nil)
      g.Assert(enrollment.UserID).Equal(int64(112))
      g.Assert(enrollment.GroupID).Equal(int64(1))

      // admin
      w = tape.PostWithClaims("/api/v1/courses/1/groups/2/enrollments", H{"user_id": 112}, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      enrollment, err = stores.Group.GetGroupEnrollmentOfUserInCourse(112, 1)
      g.Assert(err).Equal(nil)
      g.Assert(enrollment.UserID).Equal(int64(112))
      g.Assert(enrollment.GroupID).Equal(int64(2))

    })

    g.It("should manually add user to group (create)", func() {

      // remove all user_group from student
      _, err := tape.DB.Exec("DELETE FROM user_group WHERE user_id = 112;")
      g.Assert(err).Equal(nil)

      w := tape.Post("/api/v1/courses/1/groups/1/enrollments", H{"user_id": 112})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // student
      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/enrollments", H{"user_id": 112}, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutor
      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/enrollments", H{"user_id": 112}, 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PostWithClaims("/api/v1/courses/1/groups/1/enrollments", H{"user_id": 112}, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      enrollment, err := stores.Group.GetGroupEnrollmentOfUserInCourse(112, 1)
      g.Assert(err).Equal(nil)
      g.Assert(enrollment.UserID).Equal(int64(112))
      g.Assert(enrollment.GroupID).Equal(int64(1))

      // admin
      w = tape.PostWithClaims("/api/v1/courses/1/groups/2/enrollments", H{"user_id": 112}, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      enrollment, err = stores.Group.GetGroupEnrollmentOfUserInCourse(112, 1)
      g.Assert(err).Equal(nil)
      g.Assert(enrollment.UserID).Equal(int64(112))
      g.Assert(enrollment.GroupID).Equal(int64(2))

    })

    g.It("Should be able to filter enrollments (all)", func() {
      group_active, err := stores.Group.Get(1)
      g.Assert(err).Equal(nil)

      number_enrollments_expected, err := countEnrollments(
        tape,
        "SELECT count(*) FROM user_group WHERE group_id = $1",
        group_active.ID,
      )
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/groups/1/enrollments", 1, true)
      enrollments_actual := []enrollmentResponse{}
      err = json.NewDecoder(w.Body).Decode(&enrollments_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(enrollments_actual)).Equal(number_enrollments_expected)
    })

    g.AfterEach(func() {
      tape.AfterEach()
    })

  })

}
