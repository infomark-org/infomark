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

package auth

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword uses bcrypt to securely hash a plain password
func HashPassword(plain_password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain_password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash tests whether a given plain_password matches the securely
// hashed one.
func CheckPasswordHash(plain_password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain_password))
	return err == nil
}

// GenerateToken generates a random string with a specific length
func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
