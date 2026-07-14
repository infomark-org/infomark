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

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword uses bcrypt to securely hash a plain password
func HashPassword(plainPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash tests whether a given plainPassword matches the securely
// hashed one.
func CheckPasswordHash(plainPassword, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))
	return err == nil
}

// GenerateToken generates a random string with a specific length
func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// ConstantTimeTokenCompare reports whether an attacker-supplied token equals
// the trusted token stored server-side, without leaking information through
// timing.
//
// Email-confirmation and password-reset tokens are secrets that arrive
// straight from the request body. A naive `trusted == supplied` returns as
// soon as the first differing byte is found, so the time to reject a guess
// grows with the length of the matching prefix. An attacker who can measure
// that timing can recover the token one byte at a time, defeating its secrecy
// entirely.
//
// crypto/subtle.ConstantTimeCompare removes the early-exit, but it also
// reveals whether the two inputs have equal length (it returns 0 immediately
// for a length mismatch). Because the supplied token is attacker-controlled
// its length is arbitrary, so we first reduce both sides to a fixed-length
// SHA-256 digest. The subsequent comparison then always runs over equal-length
// inputs, leaking neither the token bytes nor the trusted token's length.
func ConstantTimeTokenCompare(trusted, supplied string) bool {
	trustedDigest := sha256.Sum256([]byte(trusted))
	suppliedDigest := sha256.Sum256([]byte(supplied))
	return subtle.ConstantTimeCompare(trustedDigest[:], suppliedDigest[:]) == 1
}
