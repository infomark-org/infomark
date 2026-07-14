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

package model

import (
	"testing"

	"github.com/franela/goblin"
)

// ratingBelowRange and ratingAboveRange bracket the valid 1..5 rating range and
// are used to check the boundary rejection of TaskRating.Validate.
const (
	ratingBelowRange = 0
	ratingLowerBound = 1
	ratingUpperBound = 5
	ratingAboveRange = 6
)

func TestModel(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("User.FullName", func() {

		// FullName must join the first and last name with a single space.
		g.It("Should join first and last name", func() {
			user := &User{FirstName: "Grace", LastName: "Hopper"}
			g.Assert(user.FullName()).Equal("Grace Hopper")
		})
	})

	g.Describe("TaskRating.Validate", func() {

		// A fully populated rating within the allowed range must validate.
		g.It("Should accept a well-formed rating", func() {
			rating := &TaskRating{UserID: 1, TaskID: 1, Rating: 3}
			g.Assert(rating.Validate()).Equal(nil)
		})

		// The lower and upper bounds of the range must both be accepted.
		g.It("Should accept the range boundaries", func() {
			lower := &TaskRating{UserID: 1, TaskID: 1, Rating: ratingLowerBound}
			g.Assert(lower.Validate()).Equal(nil)
			upper := &TaskRating{UserID: 1, TaskID: 1, Rating: ratingUpperBound}
			g.Assert(upper.Validate()).Equal(nil)
		})

		// A rating above the maximum must be rejected.
		g.It("Should reject a rating above the maximum", func() {
			rating := &TaskRating{UserID: 1, TaskID: 1, Rating: ratingAboveRange}
			g.Assert(rating.Validate() == nil).IsFalse()
		})

		// A zero rating is both below the minimum and fails the required check.
		g.It("Should reject a zero rating", func() {
			rating := &TaskRating{UserID: 1, TaskID: 1, Rating: ratingBelowRange}
			g.Assert(rating.Validate() == nil).IsFalse()
		})

		// A missing user id must be rejected because it is required.
		g.It("Should reject a missing user id", func() {
			rating := &TaskRating{UserID: 0, TaskID: 1, Rating: 3}
			g.Assert(rating.Validate() == nil).IsFalse()
		})

		// A missing task id must be rejected because it is required.
		g.It("Should reject a missing task id", func() {
			rating := &TaskRating{UserID: 1, TaskID: 0, Rating: 3}
			g.Assert(rating.Validate() == nil).IsFalse()
		})
	})

	g.Describe("TaskPoints.Validate", func() {

		// TaskPoints is a read-only join view, so validation always succeeds.
		g.It("Should always accept task points", func() {
			points := &TaskPoints{AchievablePoints: 10, AquiredPoints: 5, MaxPoints: 10, TaskID: 1}
			g.Assert(points.Validate()).Equal(nil)

			// The zero value must also validate, since there are no constraints.
			empty := &TaskPoints{}
			g.Assert(empty.Validate()).Equal(nil)
		})
	})
}
