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
)

// Hint contains information for the user
type Hint struct {
  Validator string `json:"type` // short descriptor of validator
  Message   string `json:"message`
}

// Collection of all responses from the validation process
type CheckResponses struct {
  Responses []*CheckResponse `json:"hints`
}

// A single validation response
type CheckResponse struct {
  Field string  `json:"field`
  Hints []*Hint `json:"field_hints`
}

// Rule specifies how a value should behave
type Rule interface {
  Validate(value interface{}) error
  Name() string
}

// front-facing developer structure to implement rules
type Check struct {
  Field string
  Value interface{}
  Rules []Rule
}

// Run a list of validations
func Validate(vals []Check) (*CheckResponses, error) {

  resp := &CheckResponses{}

  for _, v := range vals {
    cresp := &CheckResponse{Field: v.Field}
    for _, rule := range v.Rules {
      if err := rule.Validate(v.Value); err != nil {
        cresp.Hints = append(cresp.Hints, &Hint{
          Validator: rule.Name(),
          Message:   err.Error(),
        })
      }
    }

    if len(cresp.Hints) > 0 {
      resp.Responses = append(resp.Responses, cresp)
    }

  }

  if len(resp.Responses) == 0 {
    return resp, nil
  }

  return resp, errors.New("Check failed")
}
