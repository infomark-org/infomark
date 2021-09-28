// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
// Authors: Raphael Braun
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
	"github.com/lib/pq"
	null "gopkg.in/guregu/null.v3"
)

// A team consist just of the key of the team and is referenced by the
// enrollment to a course
type Team struct {
	ID int64 `db:"id"`
}

// Model representations of a Teams for query
type TeamRecord struct {
	ID      null.Int       `db:"id"`
	UserID  int64          `db:"user_id"`
	Members pq.StringArray `db:"members"`
}

// Model representations of a boolean for query
type BoolRecord struct {
	Bool bool `db:"bool"`
}
