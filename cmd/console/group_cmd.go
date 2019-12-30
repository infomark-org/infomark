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
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/infomark-org/infomark/configuration"
	"github.com/infomark-org/infomark/model"
	"github.com/lib/pq"
	"github.com/spf13/cobra"
)

func init() {
	GroupCmd.AddCommand(GroupReadBids)
	GroupCmd.AddCommand(GroupParseBidsSolution)
	GroupCmd.AddCommand(GroupEnroll)
	GroupCmd.AddCommand(GroupList)
	GroupCmd.AddCommand(GroupLocate)
	GroupCmd.AddCommand(GroupUserBids)

}

var GroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Management of groups",
}

type groupSummary struct {
	Count       int    `db:"count"`
	GroupID     int    `db:"group_id"`
	Description string `db:"description"`
}

var GroupLocate = &cobra.Command{
	Use:   "locate [courseID] [userID]",
	Short: "locate a student in a group",
	Long:  `show the exercise group for a given student`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		courseID := MustInt64Parameter(args[0], "courseID")
		userID := MustInt64Parameter(args[1], "userID")

		configuration.MustFindAndReadConfiguration()

		_, stores := MustConnectAndStores()
		user, err := stores.User.Get(userID)
		if err != nil {
			log.Fatalf("user with id %v not found\n", userID)
		}

		course, err := stores.Course.Get(courseID)
		if err != nil {
			log.Fatalf("course with id %v not found\n", courseID)
		}

		groups, err := stores.Group.GetInCourseWithUser(user.ID, course.ID)
		failWhenSmallestWhiff(err)

		if len(groups) == 0 {
			log.Fatalf("user %s %s (%d) is not enrolled as a student in course %s (%d)",
				user.FirstName, user.LastName, user.ID, course.Name, course.ID)
		}

		group := groups[0]

		fmt.Printf("found\n")
		fmt.Printf(" - Group (%d):   %s\n", group.ID, group.Description)
		fmt.Printf(" - Tutor (%d):   %s %s\n", group.TutorID, group.TutorFirstName, group.TutorLastName)
		fmt.Printf(" - Student (%d): %s %s \n", user.ID, user.FirstName, user.LastName)

	},
}

var GroupList = &cobra.Command{
	Use:   "list [courseID]",
	Short: "list all groups from a specific course",
	Long: `shows information about exercise groups with their description and
number of assigned students`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		courseID := MustInt64Parameter(args[0], "courseID")

		configuration.MustFindAndReadConfiguration()

		db, _ := MustConnectAndStores()

		groupSummaries := []groupSummary{}

		err := db.Select(&groupSummaries, `
SELECT
  count(*), ug.group_id, g.description
FROM
  user_group ug
INNER JOIN groups g ON g.id = ug.group_id
WHERE
  g.course_id = $1
GROUP BY
  ug.group_id, g.description
ORDER BY g.description
    `, courseID)
		failWhenSmallestWhiff(err)

		fmt.Printf("count   groupID    description\n")
		for k, v := range groupSummaries {
			fmt.Printf("%5d  %7d   %s\n", v.Count, v.GroupID, v.Description)
			if k%5 == 0 {
				fmt.Println("")
			}
		}
	},
}

var GroupEnroll = &cobra.Command{
	Use:   "enroll [groupID] [userID]",
	Short: "enroll a student to a group",
	Long: `enroll a student to a group or update enrollment if student is
already enrolled in another group`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupID := MustInt64Parameter(args[0], "groupID")
		userID := MustInt64Parameter(args[1], "userID")

		// same as POST "/courses/{course_id}/groups/{group_id}/enrollments"
		// TODO(patwie): good candidate for a remote cli

		configuration.MustFindAndReadConfiguration()

		_, stores := MustConnectAndStores()

		user, err := stores.User.Get(userID)
		failWhenSmallestWhiff(err)

		course, err := stores.Group.IdentifyCourseOfGroup(groupID)
		failWhenSmallestWhiff(err)

		enrollment, err := stores.Group.GetGroupEnrollmentOfUserInCourse(userID, course.ID)

		if err != nil {
			// does not exists yet
			enrollment := &model.GroupEnrollment{
				UserID:  userID,
				GroupID: groupID,
			}

			_, err := stores.Group.CreateGroupEnrollmentOfUserInCourse(enrollment)
			failWhenSmallestWhiff(err)

		} else {
			group, err := stores.Group.Get(enrollment.GroupID)
			failWhenSmallestWhiff(err)

			fmt.Printf("user %s %s (id: %v) was    enrolled in group (%v) %s\n",
				user.FirstName,
				user.LastName,
				user.ID,
				group.ID, group.Description)

			// does exists --> simply change it
			enrollment.GroupID = groupID
			err = stores.Group.ChangeGroupEnrollmentOfUserInCourse(enrollment)
			failWhenSmallestWhiff(err)

		}

		group, err := stores.Group.Get(groupID)
		failWhenSmallestWhiff(err)

		fmt.Printf("user %s %s (id: %v) is now enrolled in group (%v) %s\n",
			user.FirstName,
			user.LastName,
			user.ID,
			group.ID, group.Description)

	},
}

type userBidSummary struct {
	Bid         int    `db:"bid"`
	GroupID     int    `db:"group_id"`
	Description string `db:"description"`
}

var GroupUserBids = &cobra.Command{
	Use:   "bids [courseID] [userID]",
	Short: "list all bids of a user",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		courseID := MustInt64Parameter(args[0], "courseID")
		userID := MustInt64Parameter(args[1], "userID")

		// same as POST "/courses/{course_id}/groups/{group_id}/enrollments"
		// TODO(patwie): good candidate for a remote cli

		configuration.MustFindAndReadConfiguration()

		db, stores := MustConnectAndStores()

		user, err := stores.User.Get(userID)
		if err != nil {
			log.Fatalf("user with id %v not found\n", userID)
		}

		course, err := stores.Course.Get(courseID)
		if err != nil {
			log.Fatalf("course with id %v not found\n", courseID)
		}

		userBidSummaries := []userBidSummary{}

		err = db.Select(&userBidSummaries, `
SELECT
  bid, group_id, description
FROM
  group_bids gb
INNER JOIN groups g ON gb.group_id = g.id
WHERE
  gb.user_id = $1
AND
  g.course_id = $2
    `, user.ID, course.ID)
		failWhenSmallestWhiff(err)

		fmt.Printf("  bid   groupID    description\n")
		for k, v := range userBidSummaries {
			fmt.Printf("%5d  %7d   %s\n", v.Bid, v.GroupID, v.Description)
			if k%5 == 0 && k > 0 {
				fmt.Println("")
			}
		}

	},
}

var GroupReadBids = &cobra.Command{
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
		failWhenSmallestWhiff(err)
		defer f.Close()

		configuration.MustFindAndReadConfiguration()

		db, stores := MustConnectAndStores()

		course, err := stores.Course.Get(courseID)
		if err != nil {
			log.Fatalf("course with id %v not found\n", course.ID)
		}

		groups, err := stores.Group.GroupsOfCourse(course.ID)
		failWhenSmallestWhiff(err)

		fmt.Printf("found %v groups\n", len(groups))

		groupIDArray := []int64{}
		for _, group := range groups {
			groupIDArray = append(groupIDArray, group.ID)
		}
		// collect all user-ids to anonymize output
		gids := strings.Trim(strings.Replace(fmt.Sprint(groupIDArray), " ", " g", -1), "[]")

		students, err := stores.Course.EnrolledUsers(course.ID,
			[]string{"0"},
			"%%",
			"%%",
			"%%",
			"%%",
			"%%",
		)
		failWhenSmallestWhiff(err)

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
  group_id ASC`, student.ID, pq.Array(groupIDArray))
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
		failWhenSmallestWhiff(err)
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
		failWhenSmallestWhiff(err)
		fpar.WriteString(fmt.Sprintf("time_limit 50\n"))
		fpar.WriteString(fmt.Sprintf("\n"))
		defer fpar.Close()

	},
}

var GroupParseBidsSolution = &cobra.Command{

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

		configuration.MustFindAndReadConfiguration()

		db, stores := MustConnectAndStores()

		course, err := stores.Course.Get(courseID)
		if err != nil {
			log.Fatalf("course with id %v not found\n", course.ID)
		}
		fmt.Println("work on course", course.ID)

		file, err := os.Open(args[1])
		failWhenSmallestWhiff(err)
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
				failWhenSmallestWhiff(err)
				w, err := strconv.Atoi(parts[1][1:])
				failWhenSmallestWhiff(err)

				assignments = append(assignments, Assignment{GroupID: int64(w), UserID: int64(v)})
				uids = append(uids, int64(v))
			}

		}

		// we perform the update as a transaction
		tx, err := db.Begin()
		failWhenSmallestWhiff(err)
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
			failWhenSmallestWhiff(err)
		}

		for _, assignment := range assignments {
			fmt.Printf("%v %v\n", assignment.UserID, assignment.GroupID)

			_, err = tx.Exec("INSERT INTO user_group (id,user_id,group_id) VALUES (DEFAULT,$1,$2);", assignment.UserID, assignment.GroupID)
			if err != nil {
				tx.Rollback()
				failWhenSmallestWhiff(err)
			}
		}

		// run transaction
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			failWhenSmallestWhiff(err)
		}

		fmt.Println("Done")

	},
}
