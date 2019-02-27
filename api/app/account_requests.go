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

package app

import (
	"errors"
	"net/http"
	"strings"

	"github.com/cgtuebingen/infomark-backend/auth"
	"github.com/cgtuebingen/infomark-backend/model"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/spf13/viper"
	null "gopkg.in/guregu/null.v3"
)

// accountInfo is the request payload when sending a request to /account endpoint.
type accountInfo struct {
	Email             string `json:"email"`
	PlainPassword     string `json:"plain_password"`
	EncryptedPassword string `json:"-"`
}

// -----------------------------------------------------------------------------

// createUserAccountRequest is the request payload when registering a new user.
type createUserAccountRequest struct {
	User    *model.User  `json:"user"`
	Account *accountInfo `json:"account"`
}

// Bind preprocesses a createUserAccountRequest.
func (body *createUserAccountRequest) Bind(r *http.Request) (err error) {
	// sending the id via request is invalid as the id should be submitted in the url
	if body.User == nil {
		return errors.New("user data is missing")
	}

	if body.Account == nil {
		return errors.New("account data is missing")
	}

	// override ID, IDs should be within the URL
	body.User.ID = 0

	// check password length
	if len(body.Account.PlainPassword) < viper.GetInt("min_password_length") {
		return errors.New("password too short")
	}

	// encrypt password
	body.User.EncryptedPassword, err = auth.HashPassword(body.Account.PlainPassword)
	if err != nil {
		return err
	}
	body.User.AvatarURL = null.String{}

	if err := body.User.Validate(); err != nil {
		return err
	}

	body.Account.Email = strings.TrimSpace(body.Account.Email)
	body.Account.Email = strings.ToLower(body.Account.Email)

	if body.User.Email != body.Account.Email {
		return errors.New("email from user does not match email from account")
	}

	return nil
}

// -----------------------------------------------------------------------------

// accountRequest is the request payload for account management (email, password)
type accountRequest struct {
	Account          *accountInfo `json:"account"`
	OldPlainPassword string       `json:"old_plain_password"`
}

// Bind preprocesses a accountRequest.
func (body *accountRequest) Bind(r *http.Request) error {
	// this is the only patch function
	// sending the id via request is invalid as the id should be submitted in the url
	if body.Account == nil {
		return errors.New("missing account data")
	}

	body.Account.Email = strings.TrimSpace(body.Account.Email)
	body.Account.Email = strings.ToLower(body.Account.Email)

	// encrypt new password, when given
	if body.Account.PlainPassword != "" {
		hash, err := auth.HashPassword(body.Account.PlainPassword)
		body.Account.EncryptedPassword = hash
		return err
	}

	// validate email, when given
	return validation.ValidateStruct(body.Account,
		validation.Field(&body.Account.Email, is.Email),
	)

	return nil
}
