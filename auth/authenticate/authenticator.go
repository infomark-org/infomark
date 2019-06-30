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
	"net/http"

	"github.com/alexedwards/scs"
	"github.com/go-chi/jwtauth"
	"github.com/spf13/viper"
)

var SessionManager *scs.Manager

func PrepareSessionManager() {
	SessionManager = createSessionManager()
}

// createSessionManager starts a web session and stores the information into a
// http-only cookie. This is the prefered way when using a SPA.
func createSessionManager() *scs.Manager {
	sessionManager := scs.NewCookieManager(viper.GetString("auth_session_secret"))
	sessionManager.Lifetime(viper.GetDuration("auth_cookie_lifetime"))        // Set the maximum session lifetime to 1 hour.
	sessionManager.IdleTimeout(viper.GetDuration("auth_cookie_idle_timeout")) // Set the maximum session lifetime without actions.
	sessionManager.Persist(true)                                              // Persist the session after a user has closed their browser.
	sessionManager.Secure(viper.GetBool("auth_secure_cookie"))                // Set the Secure flag on the session cookie.
	return sessionManager
}

// HasHeaderToken tests if the request header has a token without verifying the
// correctness.
func HasHeaderToken(r *http.Request) bool {
	jwt := jwtauth.TokenFromHeader(r)
	return jwt != ""
}

// HasSessionToken tests if the request header has the http-only cookies
// containing session informations.
func HasSessionToken(r *http.Request) bool {
	session := SessionManager.Load(r)

	// try to extract the login_id which is the identifier of the request identity.
	loginID, err := session.GetInt64("login_id")
	if err != nil {
		return false
	}

	// ids will start from 1
	// this has been used for testing. In JWT we will allow id 0 for background workers.
	if loginID == 0 {
		return false
	}

	return true
}
