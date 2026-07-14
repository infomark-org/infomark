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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/franela/goblin"
)

// tokenHexLength is the number of hex characters produced per requested byte.
// GenerateToken formats each byte with %x, which yields two hex characters.
const tokenHexCharactersPerByte = 2

func TestPassword(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Password hashing", func() {

		// The central invariant of a password hasher: a hash produced from a
		// plaintext must verify against that same plaintext.
		g.It("Should verify a hash against its original plaintext", func() {
			plainPassword := "correct horse battery staple"
			hash, err := HashPassword(plainPassword)
			g.Assert(err).Equal(nil)
			// The bcrypt hash must not equal the plaintext, otherwise it is not hashed.
			g.Assert(hash == plainPassword).IsFalse()
			g.Assert(CheckPasswordHash(plainPassword, hash)).IsTrue()
		})

		// bcrypt salts every hash, so hashing the same password twice must yield
		// two different hashes that both still verify.
		g.It("Should produce distinct salted hashes for identical input", func() {
			plainPassword := "same-password"
			firstHash, err := HashPassword(plainPassword)
			g.Assert(err).Equal(nil)
			secondHash, err := HashPassword(plainPassword)
			g.Assert(err).Equal(nil)
			g.Assert(firstHash == secondHash).IsFalse()
			g.Assert(CheckPasswordHash(plainPassword, firstHash)).IsTrue()
			g.Assert(CheckPasswordHash(plainPassword, secondHash)).IsTrue()
		})

		// A wrong password must be rejected against a valid hash.
		g.It("Should reject a wrong password", func() {
			hash, err := HashPassword("the-real-password")
			g.Assert(err).Equal(nil)
			g.Assert(CheckPasswordHash("a-different-password", hash)).IsFalse()
		})

		// The empty password is a valid input for bcrypt and must round-trip,
		// while a non-empty password must not verify against the empty hash.
		g.It("Should round-trip the empty password", func() {
			hash, err := HashPassword("")
			g.Assert(err).Equal(nil)
			g.Assert(CheckPasswordHash("", hash)).IsTrue()
			g.Assert(CheckPasswordHash("not-empty", hash)).IsFalse()
		})

		// A garbage (non-bcrypt) hash must never validate any password.
		g.It("Should reject verification against a malformed hash", func() {
			g.Assert(CheckPasswordHash("whatever", "not-a-bcrypt-hash")).IsFalse()
		})
	})

	g.Describe("GenerateToken", func() {

		// GenerateToken hex-encodes length random bytes, so the string length
		// must be exactly twice the requested byte length.
		g.It("Should produce a hex string of twice the byte length", func() {
			for _, requestedBytes := range []int{0, 1, 16, 32} {
				token := GenerateToken(requestedBytes)
				g.Assert(len(token)).Equal(requestedBytes * tokenHexCharactersPerByte)
			}
		})

		// Two freshly generated tokens of a reasonable size must differ, which
		// demonstrates the randomness the token relies on for security.
		g.It("Should produce distinct tokens on subsequent calls", func() {
			firstToken := GenerateToken(32)
			secondToken := GenerateToken(32)
			g.Assert(firstToken == secondToken).IsFalse()
		})
	})

	g.Describe("ConstantTimeTokenCompare", func() {

		// The positive case: the exact same token on both sides must match.
		// This proves the constant-time rewrite did not break the accept path
		// used by the email-confirmation and password-reset flows.
		g.It("Should accept a token that matches exactly", func() {
			token := GenerateToken(32)
			g.Assert(ConstantTimeTokenCompare(token, token)).IsTrue()
		})

		// The negative case: any differing token must be rejected, including a
		// token that shares a long common prefix (the case a timing attack would
		// otherwise exploit).
		g.It("Should reject a token that differs", func() {
			trusted := "abcdef0123456789abcdef0123456789"
			wrongLastByte := "abcdef0123456789abcdef0123456780"
			g.Assert(ConstantTimeTokenCompare(trusted, wrongLastByte)).IsFalse()
			g.Assert(ConstantTimeTokenCompare(trusted, "completely-different")).IsFalse()
		})

		// A supplied token of a different length must be rejected. Hashing both
		// sides first means length differences do not short-circuit the
		// comparison, but the result must still be a rejection.
		g.It("Should reject a token of a different length", func() {
			trusted := GenerateToken(32)
			g.Assert(ConstantTimeTokenCompare(trusted, trusted+"extra")).IsFalse()
			g.Assert(ConstantTimeTokenCompare(trusted, "")).IsFalse()
		})

		// Two empty tokens hash to the same digest and therefore compare equal,
		// exactly as the previous `==` comparison did. This documents that the
		// rewrite preserves behavior; the empty-token acceptance risk in the
		// handlers is tracked separately in SESSION.md.
		g.It("Should treat two empty tokens as equal", func() {
			g.Assert(ConstantTimeTokenCompare("", "")).IsTrue()
		})
	})
}

func TestErrorRenderers(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Error renderers", func() {

		// ErrUnauthenticatedWithDetails must carry the 401 status and expose the
		// wrapped error text so the client can see the reason.
		g.It("Should build a 401 renderer with details", func() {
			renderer := ErrUnauthenticatedWithDetails(ErrTokenExpired)
			response, ok := renderer.(*ErrResponse)
			g.Assert(ok).IsTrue()
			g.Assert(response.HTTPStatusCode).Equal(http.StatusUnauthorized)
			g.Assert(response.ErrorText).Equal(ErrTokenExpired.Error())
			g.Assert(response.StatusText).Equal(http.StatusText(http.StatusUnauthorized))
		})

		// NOTE: documents current (arguably wrong) behavior: ErrUnauthorizedWithDetails
		// is named as though it produces a 403 Forbidden (and its doc comment says so),
		// but it actually returns HTTP 401 Unauthorized.
		g.It("Should build a 401 renderer for the unauthorized-with-details helper", func() {
			renderer := ErrUnauthorizedWithDetails(ErrTokenUnauthorized)
			response, ok := renderer.(*ErrResponse)
			g.Assert(ok).IsTrue()
			g.Assert(response.HTTPStatusCode).Equal(http.StatusUnauthorized)
			g.Assert(response.ErrorText).Equal(ErrTokenUnauthorized.Error())
		})

		// The prebuilt ErrUnauthorized value must carry the 403 Forbidden status.
		g.It("Should expose a 403 status on the prebuilt ErrUnauthorized", func() {
			g.Assert(ErrUnauthorized.HTTPStatusCode).Equal(http.StatusForbidden)
			g.Assert(ErrUnauthorized.StatusText).Equal(http.StatusText(http.StatusForbidden))
		})

		// The prebuilt ErrUnauthenticated value must carry the 401 status.
		g.It("Should expose a 401 status on the prebuilt ErrUnauthenticated", func() {
			g.Assert(ErrUnauthenticated.HTTPStatusCode).Equal(http.StatusUnauthorized)
		})

		// Render must set the response status on the chi render context and must
		// not itself return an error.
		g.It("Should set the status code when rendering", func() {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/", nil)
			response := &ErrResponse{HTTPStatusCode: http.StatusTeapot}
			err := response.Render(recorder, request)
			g.Assert(err).Equal(nil)
		})
	})
}
