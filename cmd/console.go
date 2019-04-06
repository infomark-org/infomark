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

package cmd

import (
  "bufio"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "os"
  "strconv"
  "strings"

  "github.com/cgtuebingen/infomark-backend/api/app"
  "github.com/cgtuebingen/infomark-backend/api/shared"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/cgtuebingen/infomark-backend/service"
  "github.com/go-ozzo/ozzo-validation/is"
  "github.com/jmoiron/sqlx"
  "github.com/lib/pq"
  "github.com/spf13/cobra"
  "github.com/spf13/viper"
  null "gopkg.in/guregu/null.v3"
)

func fail(err error) {
  if err != nil {
    panic(fail)
  }
}

var ConsoleCmd = &cobra.Command{
  Use:   "console",
  Short: "infomark console commands",
}

func ConnectAndStores() (*sqlx.DB, *app.Stores, error) {

  db, err := sqlx.Connect("postgres", viper.GetString("database_connection"))
  if err != nil {
    return nil, nil, err
  }

  if err := db.Ping(); err != nil {
    return nil, nil, err
  }

  stores := app.NewStores(db)

  return db, stores, nil

}

var AdminAddCmd = &cobra.Command{
  Use:   "add [userID]",
  Short: "set gives an user global admin permission",
  Long:  `Will set the gobal root flag o true for a user bypassing all permission tests`,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }
    userID := int64(arg0)

    _, stores, err := ConnectAndStores()
    fail(err)

    user, err := stores.User.Get(userID)
    if err != nil {
      fmt.Printf("user with id %v not found\n", userID)
      return
    }
    user.Root = true
    stores.User.Update(user)

    fmt.Printf("user %s %s (%v) is now an admin\n", user.FirstName, user.LastName, user.ID)

  },
}

var AdminRemoveCmd = &cobra.Command{
  Use:   "remove [userID]",
  Short: "removes global admin permission from a user",
  Long:  `Will set the gobal root flag to false for a user `,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }
    userID := int64(arg0)

    _, stores, err := ConnectAndStores()
    fail(err)

    user, err := stores.User.Get(userID)
    if err != nil {
      fmt.Printf("user with id %v not found\n", userID)
      return
    }
    user.Root = false
    stores.User.Update(user)

    fmt.Printf("user %s %s (%v) is not an admin anymore\n", user.FirstName, user.LastName, user.ID)

  },
}

var UserFindCmd = &cobra.Command{
  Use:   "find [query]",
  Short: "find user by first_name, last_name or email",
  Long:  `List all users matching the query`,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    db, _, err := ConnectAndStores()
    fail(err)

    query := fmt.Sprintf("%%%s%%", args[0])

    users := []model.User{}
    err = db.Select(&users, `
    SELECT *
FROM users
WHERE
 last_name LIKE $1
OR
 first_name LIKE $1
OR
 email LIKE $1`, query)
    if err != nil {
      panic(err)
    }

    fmt.Printf("found %v users matching %s\n", len(users), query)
    for k, user := range users {
      fmt.Printf("%4d %20s %20s %50s\n", user.ID, user.FirstName, user.LastName, user.Email)
      if k%10 == 0 && k != 0 {
        fmt.Println("")
      }
    }

    fmt.Printf("found %v users matching %s\n", len(users), query)
  },
}

var UserConfirmCmd = &cobra.Command{
  Use:   "confirm [email]",
  Short: "confirms the email address manually",
  Long:  `Will run confirmation procedure for an user `,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    _, stores, err := ConnectAndStores()
    fail(err)

    email := args[0]

    user, err := stores.User.FindByEmail(email)
    if err != nil {
      fmt.Printf("user with email %v not found\n", email)
      return
    }
    user.ConfirmEmailToken = null.String{}
    stores.User.Update(user)

    fmt.Printf("email of user %s %s has been confirmed", user.FirstName, user.LastName)
  },
}

var UserSetEmailCmd = &cobra.Command{
  Use:   "set-email [userID] [email]",
  Short: "will alter the email address",
  Long:  `Will change email address of an user without confirmation procedure`,
  Args:  cobra.ExactArgs(2),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }
    userID := int64(arg0)
    email := args[1]

    err = is.Email.Validate(email)
    if err != nil {
      fmt.Printf("email '%s' is not a valid email\n", email)
      return
    }

    _, stores, err := ConnectAndStores()
    fail(err)

    user, err := stores.User.Get(userID)
    if err != nil {
      fmt.Printf("user with id %v not found\n", userID)
      return
    }
    user.Email = email
    stores.User.Update(user)

    fmt.Printf("email of user %s %s is now %s", user.FirstName, user.LastName, user.Email)
  },
}

var UserCmd = &cobra.Command{
  Use:   "user",
  Short: "Management of users",
}

var AdminCmd = &cobra.Command{
  Use:   "admin",
  Short: "Management of admins",
}

// -----------------------------------------------------------------------------

// go build infomark.go && ./infomark console submission enqueue 24 10 24 "test_java_submission:v1"
var SubmissionEnqueueCmd = &cobra.Command{

  // ./infomark console submission enqueue 24 10 24 "test_java_submission:v1"
  // cp files/fixtures/unittest.zip files/uploads/tasks/24-public.zip
  // cp files/fixtures/submission.zip files/uploads/submissions/10.zip

  Use:   "enqueue [taskID] [submissionID] [gradeID] [dockerimage]",
  Short: "put submission into testing queue",
  Long:  `will enqueue a submission again into the testing queue`,
  Args:  cobra.ExactArgs(4),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }

    arg1, err := strconv.Atoi(args[1])
    if err != nil {
      fmt.Printf("cannot convert submissionID '%s' to int\n", args[1])
      return
    }

    arg2, err := strconv.Atoi(args[2])
    if err != nil {
      fmt.Printf("cannot convert gradeID '%s' to int\n", args[2])
      return
    }

    taskID := int64(arg0)
    submissionID := int64(arg1)
    gradeID := int64(arg2)
    dockerimage := args[3]

    log.Println("starting producer...")

    cfg := &service.Config{
      Connection:   viper.GetString("rabbitmq_connection"),
      Exchange:     viper.GetString("rabbitmq_exchange"),
      ExchangeType: viper.GetString("rabbitmq_exchangeType"),
      Queue:        viper.GetString("rabbitmq_queue"),
      Key:          viper.GetString("rabbitmq_key"),
      Tag:          "SimpleSubmission",
    }

    tokenManager, err := authenticate.NewTokenAuth()
    fail(err)
    accessToken, err := tokenManager.CreateAccessJWT(
      authenticate.NewAccessClaims(1, true))
    fail(err)

    request := &shared.SubmissionAMQPWorkerRequest{
      SubmissionID: submissionID,
      AccessToken:  accessToken,
      FrameworkFileURL: fmt.Sprintf("%s/api/v1/tasks/%s/private_file",
        viper.GetString("url"),
        strconv.FormatInt(taskID, 10)),
      SubmissionFileURL: fmt.Sprintf("%s/api/v1/submissions/%s/file",
        viper.GetString("url"),
        strconv.FormatInt(submissionID, 10)),
      ResultEndpointURL: fmt.Sprintf("%s/api/v1/grades/%s/private_result",
        viper.GetString("url"),
        strconv.FormatInt(gradeID, 10)),
      DockerImage: dockerimage,
    }

    body, err := json.Marshal(request)
    if err != nil {
      fmt.Errorf("json.Marshal: %s", err)
    }

    producer, _ := service.NewProducer(cfg)
    producer.Publish(body)

  },
}

var SubmissionCmd = &cobra.Command{
  Use:   "submission",
  Short: "Management of submission",
}

var CourseReadBids = &cobra.Command{

  // ./infomark console submission enqueue 24 10 24 "test_java_submission:v1"
  // cp files/fixtures/unittest.zip files/uploads/tasks/24-public.zip
  // cp files/fixtures/submission.zip files/uploads/submissions/10.zip

  Use:   "dump-bids [courseID] [file]  [min_per_group] [max_per_group]",
  Short: "export group-bids of all users in a course",
  Long:  `for assignment`,
  Args:  cobra.ExactArgs(4),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert courseID '%s' to int\n", args[0])
      return
    }

    minPerGroup, err := strconv.Atoi(args[2])
    if err != nil {
      fmt.Printf("cannot convert minPerGroup '%s' to int\n", args[2])
      return
    }

    maxPerGroup, err := strconv.Atoi(args[3])
    if err != nil {
      fmt.Printf("cannot convert maxPerGroup '%s' to int\n", args[3])
      return
    }

    fmt.Printf("bound Group Capacitiy >= %v\n", minPerGroup)
    fmt.Printf("bound Group Capacitiy <= %v\n", maxPerGroup)

    f, err := os.Create(fmt.Sprintf("%s.dat", args[1]))
    fail(err)
    defer f.Close()

    db, stores, err := ConnectAndStores()
    fail(err)

    course, err := stores.Course.Get(int64(arg0))
    if err != nil {
      fmt.Printf("course with id %v not found\n", course.ID)
      return
    }

    groups, err := stores.Group.GroupsOfCourse(course.ID)
    fail(err)

    fmt.Printf("found %v groups\n", len(groups))

    group_id_array := []int64{}
    for _, group := range groups {
      group_id_array = append(group_id_array, group.ID)
    }
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

    // for _, student := range students {
    //   fmt.Println(student)
    // }

    uids := ""
    for _, student := range students {
      g := strconv.FormatInt(student.ID, 10)
      uids = uids + " u" + g
    }

    // Select gb.* from group_bids gb
    // INNER JOIN groups g ON g.id = gb.group_id
    // WHERE user_id = 112
    // AND g.course_id = 1

    f.WriteString(fmt.Sprintf("set group := g%s;\n", gids))
    f.WriteString(fmt.Sprintf("set student := %s;\n", uids))
    f.WriteString(fmt.Sprintf("\n"))
    f.WriteString(fmt.Sprintf("param pref:\n"))
    f.WriteString(fmt.Sprintf("     g%s:=\n", gids))

    for _, student := range students {
      uid := strconv.FormatInt(student.ID, 10)
      // uids = uids + " uid" + uid
      bids := []model.GroupBid{}

      err := db.Select(&bids, `
    Select * from group_bids
WHERE user_id = $1
AND group_id = ANY($2)
ORDER BY group_id ASC`, student.ID, pq.Array(group_id_array))
      if err != nil {
        panic(err)
      }

      f.WriteString(fmt.Sprintf("u%s", uid))

      for _, group := range groups {
        bidValue := 0

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

    fmt.Println("run then")
    fmt.Println("")
    fmt.Println("")
    fmt.Printf("sudo docker run -v \"$PWD\":/data -it patwie/symphony  /var/symphony/bin/symphony -F %s.mod -D %s.dat -f %s.par\n", args[1], args[1], args[1])
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

  Use:   "import-solution [courseID] [file]",
  Short: "parse solution and assign students to groups",
  Long:  `for assignment`,
  Args:  cobra.ExactArgs(2),
  Run: func(cmd *cobra.Command, args []string) {

    type Assignment struct {
      GroupID int64
      UserID  int64
    }

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert courseID '%s' to int\n", args[0])
      return
    }

    db, stores, err := ConnectAndStores()
    fail(err)
    _ = db

    course, err := stores.Course.Get(int64(arg0))
    if err != nil {
      fmt.Printf("course with id %v not found\n", course.ID)
      return
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

    tx, err := db.Begin()
    // delete assignments so far
    _, err = tx.Exec(`DELETE FROM user_group ug
USING groups g
WHERE ug.group_id = g.id
AND g.course_id = $1
AND ug.user_id = ANY($2)`, course.ID, pq.Array(uids))
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

    err = tx.Commit()
    if err != nil {
      tx.Rollback()
      fail(err)
    }

    fmt.Println("Done")

  },
}

var AssignmentCmd = &cobra.Command{
  Use:   "assignments",
  Short: "Management of groups assignment",
}

func init() {

  AdminCmd.AddCommand(AdminRemoveCmd)
  AdminCmd.AddCommand(AdminAddCmd)
  ConsoleCmd.AddCommand(AdminCmd)

  UserCmd.AddCommand(UserSetEmailCmd)
  UserCmd.AddCommand(UserConfirmCmd)
  UserCmd.AddCommand(UserFindCmd)
  ConsoleCmd.AddCommand(UserCmd)

  SubmissionCmd.AddCommand(SubmissionEnqueueCmd)
  ConsoleCmd.AddCommand(SubmissionCmd)

  AssignmentCmd.AddCommand(CourseReadBids)
  AssignmentCmd.AddCommand(CourseParseBidsSolution)
  ConsoleCmd.AddCommand(AssignmentCmd)

  RootCmd.AddCommand(ConsoleCmd)
}
