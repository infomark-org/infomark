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

package validation

import (
  "errors"
  "fmt"
  "regexp"
  "unicode/utf8"
)

// Check whether the length of a given string is within the bounds.
type LengthRule struct {
  Min, Max int
}

func (v *LengthRule) Name() string {
  return "length"
}

func (v *LengthRule) Validate(value interface{}) error {
  var l int
  if s, ok := value.(string); ok {
    l = utf8.RuneCountInString(s)
  }

  if v.Min > l {
    return errors.New(fmt.Sprintf("Length (%v) too short", l))
  }
  if v.Max < l {
    return errors.New(fmt.Sprintf("Length (%v) too long", l))
  }

  return nil
}

// .............................................................................

// Check whether the string matches a regex.
type MatchRule struct {
  Expr *regexp.Regexp
}

func (v *MatchRule) Name() string {
  return "Match"
}

func (v *MatchRule) Validate(value interface{}) error {

  if s, ok := value.(string); ok {

    if v.Expr.MatchString(s) {
      return nil
    } else {
      return errors.New(fmt.Sprintf("Input does not match regex %v", v.Expr.String()))
    }

  }
  return errors.New("no string")
}
