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

package authenticate

import (
  "time"

  jwt "github.com/dgrijalva/jwt-go"
  "github.com/go-pg/pg/orm"
)

// Token holds refresh jwt information.
type Token struct {
  ID        int       `json:"id,omitempty"`
  CreatedAt time.Time `json:"created_at,omitempty"`
  UpdatedAt time.Time `json:"updated_at,omitempty"`
  AccountID int       `json:"-"`

  Token      string    `json:"-"`
  Expiry     time.Time `json:"-"`
  Mobile     bool      `sql:",notnull" json:"mobile"`
  Identifier string    `json:"identifier,omitempty"`
}

// BeforeInsert hook executed before database insert operation.
func (t *Token) BeforeInsert(db orm.DB) error {
  now := time.Now()
  if t.CreatedAt.IsZero() {
    t.CreatedAt = now
    t.UpdatedAt = now
  }
  return nil
}

// BeforeUpdate hook executed before database update operation.
func (t *Token) BeforeUpdate(db orm.DB) error {
  t.UpdatedAt = time.Now()
  return nil
}

// Claims returns the token claims to be signed
func (t *Token) Claims() jwt.MapClaims {
  return jwt.MapClaims{
    "id":    t.ID,
    "token": t.Token,
  }
}
