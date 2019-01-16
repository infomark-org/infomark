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

package store

import (
	"strconv"

	"github.com/cgtuebingen/infomark-backend/model"
)

func (ds *datastore) CountUsers() (int, error) {
	var count int
	err := ORM().Table("users").Count(&count).Error
	return count, err
}

// GetUserFromIdString retrieves the user from the database if exists
func (ds *datastore) GetUserFromIdString(userID string) (user *model.User, err error) {
	var uid int
	user = &model.User{}

	if userID != "" {
		if uid, err = strconv.Atoi(userID); err == nil {
			err = ORM().First(&user, uid).Error
		}
	}

	return user, err
}

func (ds *datastore) GetUserFromEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := ORM().Where("email = ?", email).First(&user).Error
	return user, err

}
