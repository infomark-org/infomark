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

package database

import (
	"github.com/infomark-org/infomark/model"
	"github.com/jmoiron/sqlx"
)

// MaterialStore is the store for materials (slides, additional material) for a
// lecture.
type MaterialStore struct {
	db *sqlx.DB
}

// NewMaterialStore creates a new material store.
func NewMaterialStore(db *sqlx.DB) *MaterialStore {
	return &MaterialStore{
		db: db,
	}
}

// GetAll returns all materials in the database.
func (s *MaterialStore) GetAll() ([]model.Material, error) {
	p := []model.Material{}
	err := s.db.Select(&p, "SELECT * FROM materials;")
	return p, err
}

// Get returns a material for a given id.
func (s *MaterialStore) Get(materialID int64) (*model.Material, error) {
	p := model.Material{ID: materialID}
	err := s.db.Get(&p, "SELECT * FROM materials WHERE id = $1 LIMIT 1;", p.ID)
	return &p, err
}

// Create creates a material for a given course.
func (s *MaterialStore) Create(p *model.Material, courseID int64) (*model.Material, error) {

	newID, err := Insert(s.db, "materials", p)
	if err != nil {
		return nil, err
	}

	// now associate sheet with course
	_, err = s.db.Exec(`
INSERT INTO
  material_course (id,material_id,course_id)
VALUES
  (DEFAULT, $1, $2);`,
		newID, courseID)
	if err != nil {
		return nil, err
	}

	return s.Get(newID)
}

// Update updates a given material. The id of the material should be in the material model.
func (s *MaterialStore) Update(p *model.Material) error {
	return Update(s.db, "materials", p.ID, p)
}

// Delete deletes a material for a given id.
func (s *MaterialStore) Delete(materialID int64) error {
	return Delete(s.db, "materials", materialID)
}

// MaterialsOfCourse returns all materials for a given course, which are visible for the given role.
func (s *MaterialStore) MaterialsOfCourse(courseID int64, givenRole int) ([]model.Material, error) {
	p := []model.Material{}

	err := s.db.Select(&p, `
SELECT
  m.*
FROM
  materials m
INNER JOIN material_course mc ON m.id = mc.material_id
WHERE
  mc.course_id = $1
AND
  m.required_role <= $2
ORDER BY
  m.lecture_at ASC;`, courseID, givenRole)
	return p, err
}

// IdentifyCourseOfMaterial returns the course, which is associated with the material.
func (s *MaterialStore) IdentifyCourseOfMaterial(sheetID int64) (*model.Course, error) {

	course := &model.Course{}
	err := s.db.Get(course,
		`
SELECT
  c.*
FROM
  courses c
INNER JOIN material_course mc ON mc.course_id = c.id
WHERE
  mc.material_id = $1
LIMIT 1`,
		sheetID)
	if err != nil {
		return nil, err
	}

	return course, err
}
