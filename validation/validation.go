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
  "log"
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
        log.Println("debug value ", v.Value)
        log.Println("not passed", rule)
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
