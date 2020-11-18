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
	"time"
)

// Sheet is a database entity representing an entire exercise sheet consisting
// of tasks.
type Sheet struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`

	Name      string    `db:"name"`
	PublishAt time.Time `db:"publish_at"`
	DueAt     time.Time `db:"due_at"`
}

// SheetPoints contains the performance of a specific student
type SheetPoints struct {
	AquiredPoints    int `db:"acquired_points"`
	AchievablePoints int `db:"achievable_points"`
	MaxPoints        int `db:"max_points"`
	SheetID          int `db:"sheet_id"`
}
