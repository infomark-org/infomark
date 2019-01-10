// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

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
