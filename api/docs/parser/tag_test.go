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

package parser

import (
  "testing"
)

func testTag(t *testing.T) {

  tables := []struct {
    Input            string
    ExpectedName     string
    ExpectedExample  string
    ExpectedRequired bool
  }{
    {`#json:"foo"#`, "foo", "", false},
    {`#json:"foo" example:"bar"#`, "foo", "bar", false},
    {`#json:"foo" example:"bar" required:""#`, "foo", "bar", true},
    {`#json:"foo" example:"bar" required:"true"#`, "foo", "bar", true},
    {`#json:"foo" example:"bar" required:"false"#`, "foo", "bar", false},
  }

  for _, table := range tables {

    field, err := parseTag(table.Input)
    if err != nil {
      t.Fatal(err)
    }

    got := field.Name
    want := table.ExpectedName

    if got != want {
      t.Errorf("got: %s, want: %s.", got, want)
    }

    got = field.Example
    want = table.ExpectedExample

    if got != want {
      t.Errorf("got: %s, want: %s.", got, want)
    }

    got2 := field.Required
    want2 := table.ExpectedRequired

    if got2 != want2 {
      t.Errorf("got: %v, want: %v.", got2, want2)
    }
  }

}
