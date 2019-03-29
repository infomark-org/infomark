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
  "context"
  "net/http"

  "github.com/cgtuebingen/infomark-backend/auth"
  "github.com/go-chi/jwtauth"
  "github.com/go-chi/render"
)

// RequiredValidAccessClaimsMiddleware tries to get information about the identity which
// issues a request by looking into the authorization header and then into
// the cookie.
func RequiredValidAccessClaims(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

    accessClaims := &AccessClaims{}

    // first we test the JWT autorization
    if HasHeaderToken(r) {

      // parse token from from header
      tokenStr := jwtauth.TokenFromHeader(r)

      // ok, there is a access token in the header
      err := accessClaims.ParseAccessClaimsFromToken(tokenStr)
      if err != nil {
        // fmt.Println(err)
        render.Render(w, r, auth.ErrUnauthorized)
        return
      }

    } else {
      // fmt.Println("no token, try session")
      if HasSessionToken(r) {
        // fmt.Println("found session")

        // session data is stored in cookie
        err := accessClaims.ParseRefreshClaimsFromSession(r)
        if err != nil {
          // fmt.Println(err)
          render.Render(w, r, auth.ErrUnauthorized)
          return
        }

        // session is valid --> we will extend the session
        w = accessClaims.WriteToSession(w, r)
      } else {
        // fmt.Println("NO session found")

        render.Render(w, r, auth.ErrUnauthenticated)
        return

      }

    }

    // nothing given
    // serve next
    ctx := context.WithValue(r.Context(), "access_claims", accessClaims)
    next.ServeHTTP(w, r.WithContext(ctx))
    return

  })
}
