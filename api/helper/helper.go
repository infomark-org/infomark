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

package helper

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"strings"
)

// StringArrayToIntArray converts a list of strings into a list of int or failes
func StringArrayToIntArray(values []string) ([]int, error) {
	out := make([]int, len(values))
	for index, value := range values {
		v, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		out[index] = v
	}
	return out, nil
}

// StringArrayFromURL will read an URL parameter like /api/?some_strings=foo,bar
func StringArrayFromURL(r *http.Request, name string, standard []string) []string {
	str := r.FormValue(name)
	if str != "" {
		return strings.Split(str, ",")
	}
	return standard
}

// StringFromURL will read an URL parameter like /api/?some_string=foo
func StringFromURL(r *http.Request, name string, standard string) string {
	str := r.FormValue(name)
	if str != "" {
		return str
	}
	return standard
}

// IntFromURL will read an URL parameter like /api/?some_int=3
func IntFromURL(r *http.Request, name string, standard int) int {
	str := r.FormValue(name)
	if str == "" {
		return standard
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		return standard
	}
	return i
}

// Int64FromURL will read an URL parameter like /api/?some_int=3
func Int64FromURL(r *http.Request, name string, standard int64) int64 {
	str := r.FormValue(name)
	if str == "" {
		return standard
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return standard
	}
	return int64(i)
}

// H is a neat alias
type H map[string]interface{}

// ToH converts any object into an typeless object (used by unit tests).
func ToH(z interface{}) map[string]interface{} {
	data, _ := json.Marshal(z)
	var msgMapTemplate interface{}
	_ = json.Unmarshal(data, &msgMapTemplate)
	return msgMapTemplate.(map[string]interface{})
}

// Time returns time.Now() but without nanseconds for passing unit-tests.
// There are some issues with storing and retriebing the nanoseconds.
func Time(t time.Time) time.Time {
	format := "2006-01-02 15:04:05 +0000 CET"
	R, _ := time.Parse(format, t.Format(format))
	return R
}
