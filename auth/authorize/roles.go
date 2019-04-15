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
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/go-chi/render"
)

type CourseRole int32

const (
  NOCOURSEROLE CourseRole = -1
  STUDENT      CourseRole = 0
  TUTOR        CourseRole = 1
  ADMIN        CourseRole = 2
)

func (r CourseRole) ToInt() int {
  switch r {
  default:
    return -1
  case NOCOURSEROLE:
    return -1
  case STUDENT:
    return 0
  case TUTOR:
    return 1
  case ADMIN:
    return 2
  }
}

// RequiresRole middleware restricts access to accounts having role parameter in their jwt claims.
func RequiresAtLeastCourseRole(requiredRole CourseRole) func(next http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    hfn := func(w http.ResponseWriter, r *http.Request) {
      if HasAtLeastRole(requiredRole, r) {
        next.ServeHTTP(w, r)
      } else {
        render.Render(w, r, auth.ErrUnauthorized)
      }
    }
    return http.HandlerFunc(hfn)
  }
}

func HasAtLeastRole(requiredRole CourseRole, r *http.Request) bool {
  // global root can lever out this check
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  if accessClaims.Root {
    // oh dear, sorry to ask. Please pass this check
    return true
  }

  givenRole, ok := r.Context().Value("course_role").(CourseRole)
  if !ok {
    return false
  }

  if givenRole < requiredRole {
    return false
  }

  return true
}

func EndpointRequiresRole(endpoint http.HandlerFunc, requiredRole CourseRole) http.HandlerFunc {

  fn := func(w http.ResponseWriter, r *http.Request) {

    if HasAtLeastRole(requiredRole, r) {
      endpoint.ServeHTTP(w, r)
    } else {
      render.Render(w, r, auth.ErrUnauthorized)
    }
  }

  return http.HandlerFunc(fn)

}
