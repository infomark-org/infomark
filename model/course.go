// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

package model

import (
	"time"

	"github.com/cgtuebingen/infomark-backend/validation"
)

// Course represents a course where students can enroll
type Course struct {
	// the id for this user.
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`

	Title       string    `json:"title"`
	SubTitle    string    `json:"sub_title"`
	Description string    `json:"description"`
	BeginAt     time.Time `json:"begin_at"`
	EndsAt      time.Time `json:"ends_at"`

	AuthorID int `json:"author_id"` // the one who created this course
}

// Validate validates the required fields and formats.
func (u *Course) Validate() (*validation.CheckResponses, error) {

	vals := []validation.Check{
		{
			Field: "title",
			Value: u.Title,
			Rules: []validation.Rule{
				&validation.LengthRule{Min: 1, Max: 250},
			},
		},
	}

	return validation.Validate(vals)
}
