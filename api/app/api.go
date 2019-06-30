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
	"github.com/cgtuebingen/infomark-backend/auth/authorize"
	"github.com/cgtuebingen/infomark-backend/database"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/symbol"
	"github.com/jmoiron/sqlx"
)

// UserStore defines user related database queries
type UserStore interface {
	Get(userID int64) (*model.User, error)
	Update(p *model.User) error
	GetAll() ([]model.User, error)
	Create(p *model.User) (*model.User, error)
	Delete(userID int64) error
	FindByEmail(email string) (*model.User, error)
	Find(query string) ([]model.User, error)
	GetEnrollments(userID int64) ([]model.Enrollment, error)
}

// CourseStore defines course related database queries
type CourseStore interface {
	Get(courseID int64) (*model.Course, error)
	Update(p *model.Course) error
	GetAll() ([]model.Course, error)
	Create(p *model.Course) (*model.Course, error)
	Delete(courseID int64) error
	Enroll(courseID int64, userID int64, role int64) error
	Disenroll(courseID int64, userID int64) error
	EnrolledUsers(
		courseID int64,
		roleFilter []string,
		filterFirstName string,
		filterLastName string,
		filterEmail string,
		filterSubject string,
		filterLanguage string) ([]model.UserCourse, error)
	FindEnrolledUsers(
		courseID int64,
		roleFilter []string,
		filterQuery string,
	) ([]model.UserCourse, error)
	GetUserEnrollment(courseID int64, userID int64) (*model.UserCourse, error)
	PointsForUser(userID int64, courseID int64) ([]model.SheetPoints, error)
	RoleInCourse(userID int64, courseID int64) (authorize.CourseRole, error)
	UpdateRole(courseID, userID int64, role int) error
}

// SheetStore specifies required database queries for Sheet management.
type SheetStore interface {
	Get(SheetID int64) (*model.Sheet, error)
	Update(p *model.Sheet) error
	GetAll() ([]model.Sheet, error)
	Create(p *model.Sheet, courseID int64) (*model.Sheet, error)
	Delete(SheetID int64) error
	SheetsOfCourse(courseID int64) ([]model.Sheet, error)
	IdentifyCourseOfSheet(sheetID int64) (*model.Course, error)
	PointsForUser(userID int64, sheetID int64) ([]model.TaskPoints, error)
}

// TaskStore specifies required database queries for Task management.
type TaskStore interface {
	Get(TaskID int64) (*model.Task, error)
	Update(p *model.Task) error
	GetAll() ([]model.Task, error)
	Create(p *model.Task, sheetID int64) (*model.Task, error)
	Delete(TaskID int64) error
	TasksOfSheet(sheetID int64) ([]model.Task, error)
	IdentifyCourseOfTask(taskID int64) (*model.Course, error)
	IdentifySheetOfTask(taskID int64) (*model.Sheet, error)

	GetAverageRating(taskID int64) (float32, error)
	GetRatingOfTaskByUser(taskID int64, userID int64) (*model.TaskRating, error)
	GetRating(taskRatingID int64) (*model.TaskRating, error)
	CreateRating(p *model.TaskRating) (*model.TaskRating, error)
	UpdateRating(p *model.TaskRating) error
	GetAllMissingTasksForUser(userID int64) ([]model.MissingTask, error)
}

// GroupStore specifies required database queries for Task management.
type GroupStore interface {
	Get(groupID int64) (*model.Group, error)
	GetAll() ([]model.Group, error)
	Create(p *model.Group) (*model.Group, error)
	Update(p *model.Group) error
	Delete(taskID int64) error
	// GroupsOfCourse(courseID int64) ([]model.Group, error)
	GroupsOfCourse(courseID int64) ([]model.GroupWithTutor, error)
	GetInCourseWithUser(userID int64, courseID int64) ([]model.GroupWithTutor, error)
	GetMembers(groupID int64) ([]model.User, error)
	GetOfTutor(tutorID int64, courseID int64) ([]model.GroupWithTutor, error)
	IdentifyCourseOfGroup(groupID int64) (*model.Course, error)

	GetBidOfUserForGroup(userID int64, groupID int64) (bid int, err error)
	InsertBidOfUserForGroup(userID int64, groupID int64, bid int) (int, error)
	UpdateBidOfUserForGroup(userID int64, groupID int64, bid int) (int, error)

	GetBidsForCourseForUser(courseID int64, userID int64) ([]model.GroupBid, error)
	GetBidsForCourse(courseID int64) ([]model.GroupBid, error)

	GetGroupEnrollmentOfUserInCourse(userID int64, courseID int64) (*model.GroupEnrollment, error)
	CreateGroupEnrollmentOfUserInCourse(p *model.GroupEnrollment) (*model.GroupEnrollment, error)
	ChangeGroupEnrollmentOfUserInCourse(p *model.GroupEnrollment) error

	EnrolledUsers(courseID int64, groupID int64, roleFilter []string,
		filterFirstName string, filterLastName string, filterEmail string, filterSubject string,
		filterLanguage string) ([]model.UserCourse, error)
}

// MaterialStore defines material related database queries
type MaterialStore interface {
	Get(sheetID int64) (*model.Material, error)
	Create(p *model.Material, courseID int64) (*model.Material, error)
	Update(p *model.Material) error
	Delete(sheetID int64) error
	MaterialsOfCourse(courseID int64, requiredRole int) ([]model.Material, error)
	IdentifyCourseOfMaterial(sheetID int64) (*model.Course, error)
	GetAll() ([]model.Material, error)
}

// SubmissionStore defines submission related database queries
type SubmissionStore interface {
	Get(submissionID int64) (*model.Submission, error)
	GetByUserAndTask(userID int64, taskID int64) (*model.Submission, error)
	Create(p *model.Submission) (*model.Submission, error)
	GetFiltered(filterCourseID, filterGroupID, filterUserID, filterSheetID, filterTaskID int64) ([]model.Submission, error)
}

// GradeStore defines grades related database queries
type GradeStore interface {
	GetFiltered(
		courseID int64,
		sheetID int64,
		taskID int64,
		groupID int64,
		userID int64,
		tutorID int64,
		feedback string,
		acquiredPoints int,
		publicTestStatus int,
		privateTestStatus int,
		publicExecutationState int,
		privateExecutationState int,
	) ([]model.Grade, error)
	Get(id int64) (*model.Grade, error)
	GetForSubmission(id int64) (*model.Grade, error)
	Update(p *model.Grade) error
	IdentifyCourseOfGrade(gradeID int64) (*model.Course, error)
	GetAllMissingGrades(courseID int64, tutorID int64, groupID int64) ([]model.MissingGrade, error)
	Create(p *model.Grade) (*model.Grade, error)

	UpdatePrivateTestInfo(gradeID int64, log string, status symbol.TestingResult) error
	UpdatePublicTestInfo(gradeID int64, log string, status symbol.TestingResult) error
	IdentifyTaskOfGrade(gradeID int64) (*model.Task, error)
	GetOverviewGrades(courseID int64, groupID int64) ([]model.OverviewGrade, error)
}

// API provides application resources and handlers.
type API struct {
	User       *UserResource
	Account    *AccountResource
	Auth       *AuthResource
	Course     *CourseResource
	Sheet      *SheetResource
	Task       *TaskResource
	Group      *GroupResource
	TaskRating *TaskRatingResource
	Submission *SubmissionResource
	Material   *MaterialResource
	Grade      *GradeResource
	Common     *CommonResource
}

// Stores is the collection of stores. We use this struct to express a kind of
// hierarchy of database queries, e.g. stores.User.Get(1)
type Stores struct {
	Course     CourseStore
	User       UserStore
	Sheet      SheetStore
	Task       TaskStore
	Group      GroupStore
	Submission SubmissionStore
	Material   MaterialStore
	Grade      GradeStore
}

// NewStores build all stores and connect them to a database.
func NewStores(db *sqlx.DB) *Stores {
	return &Stores{
		Course:     database.NewCourseStore(db),
		User:       database.NewUserStore(db),
		Sheet:      database.NewSheetStore(db),
		Task:       database.NewTaskStore(db),
		Group:      database.NewGroupStore(db),
		Submission: database.NewSubmissionStore(db),
		Material:   database.NewMaterialStore(db),
		Grade:      database.NewGradeStore(db),
	}
}

// NewAPI configures and returns application API.
func NewAPI(db *sqlx.DB) (*API, error) {
	stores := NewStores(db)

	api := &API{
		Account:    NewAccountResource(stores),
		Auth:       NewAuthResource(stores),
		User:       NewUserResource(stores),
		Course:     NewCourseResource(stores),
		Sheet:      NewSheetResource(stores),
		Task:       NewTaskResource(stores),
		Group:      NewGroupResource(stores),
		TaskRating: NewTaskRatingResource(stores),
		Submission: NewSubmissionResource(stores),
		Material:   NewMaterialResource(stores),
		Grade:      NewGradeResource(stores),
		Common:     NewCommonResource(stores),
	}
	return api, nil
}
