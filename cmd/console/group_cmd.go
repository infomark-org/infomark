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
  "bufio"
  "fmt"
  "io"
  "log"
  "os"
  "strconv"
  "strings"

  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/lib/pq"
  "github.com/spf13/cobra"
)

func init() {
  GroupCmd.AddCommand(CourseReadBids)
  GroupCmd.AddCommand(CourseParseBidsSolution)

}

var GroupCmd = &cobra.Command{
  Use:   "group",
  Short: "Management of groups",
}

var CourseReadBids = &cobra.Command{
  Use:   "dump-bids [courseID] [file]  [min_per_group] [max_per_group]",
  Short: "export group-bids of all users in a course",
  Long:  `for assignment`,
  Args:  cobra.ExactArgs(4),
  Run: func(cmd *cobra.Command, args []string) {

    courseID := MustInt64Parameter(args[0], "courseID")
    minPerGroup := MustIntParameter(args[2], "minPerGroup")
    maxPerGroup := MustIntParameter(args[3], "maxPerGroup")

    if minPerGroup > maxPerGroup {
      log.Fatalf("minPerGroup %d > maxPerGroup %d is infeasible",
        minPerGroup, maxPerGroup)
    }

    fmt.Printf("bound Group Capacitiy >= %v\n", minPerGroup)
    fmt.Printf("bound Group Capacitiy <= %v\n", maxPerGroup)

    f, err := os.Create(fmt.Sprintf("%s.dat", args[1]))
    fail(err)
    defer f.Close()

    db, stores := MustConnectAndStores()

    course, err := stores.Course.Get(courseID)
    if err != nil {
      log.Fatalf("course with id %v not found\n", course.ID)
    }

    groups, err := stores.Group.GroupsOfCourse(course.ID)
    fail(err)

    fmt.Printf("found %v groups\n", len(groups))

    group_id_array := []int64{}
    for _, group := range groups {
      group_id_array = append(group_id_array, group.ID)
    }
    // collect all user-ids to anonymize output
    gids := strings.Trim(strings.Replace(fmt.Sprint(group_id_array), " ", " g", -1), "[]")

    students, err := stores.Course.EnrolledUsers(course.ID,
      []string{"0"},
      "%%",
      "%%",
      "%%",
      "%%",
      "%%",
    )
    fail(err)

    fmt.Printf("found %v students\n", len(students))
    fmt.Printf("\n")
    fmt.Printf("\n")

    // collect all user-ids to anonymize output
    uids := ""
    for _, student := range students {
      g := strconv.FormatInt(student.ID, 10)
      uids = uids + " u" + g
    }

    f.WriteString(fmt.Sprintf("set group := g%s;\n", gids))
    f.WriteString(fmt.Sprintf("set student := %s;\n", uids))
    f.WriteString(fmt.Sprintf("\n"))
    f.WriteString(fmt.Sprintf("param pref:\n"))
    f.WriteString(fmt.Sprintf("     g%s:=\n", gids))

    // this includes students without any selected preferences
    for _, student := range students {
      uid := strconv.FormatInt(student.ID, 10)
      bids := []model.GroupBid{}

      err := db.Select(&bids, `
SELECT
  *
FROM
  group_bids
WHERE
  user_id = $1
AND
  group_id = ANY($2)
ORDER BY
  group_id ASC`, student.ID, pq.Array(group_id_array))
      if err != nil {
        panic(err)
      }

      f.WriteString(fmt.Sprintf("u%s", uid))

      // make sure all students have a bid
      // students without any preference for a course are given a 5
      // 0 means no interests, 10 means absolute favourite
      // default is maximum to prefer students
      for _, group := range groups {
        bidValue := 10
        for _, bid := range bids {
          if bid.GroupID == group.ID {
            bidValue = bid.Bid
            break
          }
        }
        f.WriteString(fmt.Sprintf(" %v", bidValue))
      }
      f.WriteString(fmt.Sprintf("\n"))
    }

    f.WriteString(fmt.Sprintf(";\n"))
    f.WriteString(fmt.Sprintf("\n"))
    f.WriteString(fmt.Sprintf("end;\n"))

    // write mod
    //
    fmod, err := os.Create(fmt.Sprintf("%s.mod", args[1]))
    fail(err)
    defer fmod.Close()

    fmod.WriteString(fmt.Sprintf("set student;\n"))
    fmod.WriteString(fmt.Sprintf("set group;\n"))
    fmod.WriteString(fmt.Sprintf("\n"))
    fmod.WriteString(fmt.Sprintf("var assign{i in student, j in group} binary;\n"))
    fmod.WriteString(fmt.Sprintf("param pref{i in student, j in group};\n"))
    fmod.WriteString(fmt.Sprintf("\n"))
    fmod.WriteString(fmt.Sprintf("maximize totalPref:\n"))
    fmod.WriteString(fmt.Sprintf("    sum{i in student, j in group} pref[i,j]*assign[i,j];\n"))
    fmod.WriteString(fmt.Sprintf("\n"))
    fmod.WriteString(fmt.Sprintf("subject to exactly_one_group {i in student}:\n"))
    fmod.WriteString(fmt.Sprintf("    sum {j in group} assign[i,j] =1;\n"))
    fmod.WriteString(fmt.Sprintf("\n"))
    fmod.WriteString(fmt.Sprintf("subject to min3{j in group}:\n"))
    fmod.WriteString(fmt.Sprintf("    sum{i in student} assign[i,j]>=%v;\n", minPerGroup))
    fmod.WriteString(fmt.Sprintf("\n"))
    fmod.WriteString(fmt.Sprintf("subject to max4{j in group}:\n"))
    fmod.WriteString(fmt.Sprintf("    sum{i in student} assign[i,j]<=%v;\n", maxPerGroup))
    fmod.WriteString(fmt.Sprintf("\n"))
    fmod.WriteString(fmt.Sprintf("end;\n"))
    fmod.WriteString(fmt.Sprintf("\n"))
    fmod.WriteString(fmt.Sprintf("\n"))

    fmt.Println("run the command")
    fmt.Println("")
    fmt.Println("")
    fmt.Printf("sudo docker run -v \"$PWD\":/data -it patwie/symphony  /var/symphony/bin/symphony -F %s.mod -D /data/%s.dat -f /data/%s.par\n", args[1], args[1], args[1])
    fmt.Println("")

    fpar, err := os.Create(fmt.Sprintf("%s.par", args[1]))
    fail(err)
    fpar.WriteString(fmt.Sprintf("time_limit 50\n"))
    fpar.WriteString(fmt.Sprintf("\n"))
    defer fpar.Close()

  },
}

var CourseParseBidsSolution = &cobra.Command{

  // ./infomark console submission enqueue 24 10 24 "test_java_submission:v1"
  // cp files/fixtures/unittest.zip files/uploads/tasks/24-public.zip
  // cp files/fixtures/submission.zip files/uploads/submissions/10.zip

  Use:   "import-assignments [courseID] [file]",
  Short: "parse solution and assign students to groups",
  Long:  `for assignment`,
  Args:  cobra.ExactArgs(2),
  Run: func(cmd *cobra.Command, args []string) {

    type Assignment struct {
      GroupID int64
      UserID  int64
    }

    courseID := MustInt64Parameter(args[0], "courseID")

    db, stores := MustConnectAndStores()

    course, err := stores.Course.Get(courseID)
    if err != nil {
      log.Fatalf("course with id %v not found\n", course.ID)
    }
    fmt.Println("work on course", course.ID)

    file, err := os.Open(args[1])
    fail(err)
    defer file.Close()

    reader := bufio.NewReader(file)
    uids := []int64{}
    assignments := []Assignment{}
    for {
      line, _, err := reader.ReadLine()
      if err == io.EOF {
        break
      }
      // parse line
      string := fmt.Sprintf("%s", line)
      if strings.HasPrefix(string, "assign[") {
        parts := strings.Split(string, "[")
        parts = strings.Split(parts[1], "]")
        parts = strings.Split(parts[0], ",")

        v, err := strconv.Atoi(parts[0][1:])
        fail(err)
        w, err := strconv.Atoi(parts[1][1:])
        fail(err)

        assignments = append(assignments, Assignment{GroupID: int64(w), UserID: int64(v)})
        uids = append(uids, int64(v))
      }

    }

    // we perform the update as a transaction
    tx, err := db.Begin()
    // delete assignments so far
    _, err = tx.Exec(`
DELETE FROM
  user_group ug
USING
  groups g
WHERE
  ug.group_id = g.id
AND
  g.course_id = $1
AND
  ug.user_id = ANY($2)`, course.ID, pq.Array(uids))
    if err != nil {
      fmt.Println(err)
      tx.Rollback()
      fail(err)
    }

    for _, assignment := range assignments {
      fmt.Printf("%v %v\n", assignment.UserID, assignment.GroupID)

      _, err = tx.Exec("INSERT INTO user_group (id,user_id,group_id) VALUES (DEFAULT,$1,$2);", assignment.UserID, assignment.GroupID)
      if err != nil {
        tx.Rollback()
        fail(err)
      }
    }

    // run transaction
    err = tx.Commit()
    if err != nil {
      tx.Rollback()
      fail(err)
    }

    fmt.Println("Done")

  },
}
