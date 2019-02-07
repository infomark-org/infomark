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
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs"
	"github.com/go-chi/jwtauth"
	"github.com/spf13/viper"
)

var SessionManager = createSessionManager()

func createSessionManager() *scs.Manager {
	sessionManager := scs.NewCookieManager(viper.GetString("auth_session_secret"))
	sessionManager.Lifetime(time.Hour) // Set the maximum session lifetime to 1 hour.
	sessionManager.Persist(true)       // Persist the session after a user has closed their browser.
	sessionManager.Secure(true)        // Set the Secure flag on the session cookie.
	return sessionManager
}

func HasHeaderToken(r *http.Request) bool {
	jwt := jwtauth.TokenFromHeader(r)
	return jwt != ""
}

func HasSessionToken(r *http.Request) bool {
	session := SessionManager.Load(r)

	loginID, err := session.GetInt64("login_id")
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("loginID", loginID)

	if loginID == 0 {
		return false
	}

	return true
}
