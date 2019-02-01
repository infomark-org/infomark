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

package authorize

import (
  "net/http"

  "github.com/cgtuebingen/infomark-backend/auth"
  "github.com/go-chi/render"

  "github.com/dhax/go-base/auth/jwt"
)

// RequiresRole middleware restricts access to accounts having role parameter in their jwt claims.
func RequiresRole(role string) func(next http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    hfn := func(w http.ResponseWriter, r *http.Request) {
      claims := jwt.ClaimsFromCtx(r.Context())
      if !hasRole(role, claims.Roles) {
        render.Render(w, r, auth.ErrUnauthorized)
        return
      }
      next.ServeHTTP(w, r)
    }
    return http.HandlerFunc(hfn)
  }
}

func hasRole(role string, roles []string) bool {
  // for _, r := range roles {
  //   if r == role {
  //     return true
  //   }
  // }
  return false
}
