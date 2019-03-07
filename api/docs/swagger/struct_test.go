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

package swagger

import (
  "fmt"
  "go/ast"
  "go/parser"
  "go/token"
  "testing"
)

func TestStruct(t *testing.T) {

  fset := token.NewFileSet() // positions are relative to fset

  d, err := parser.ParseDir(fset, "./fixture", nil, parser.ParseComments)
  if err != nil {
    fmt.Println(err)
    return
  }

  // gather structs
  for _, pkg := range d {
    ast.Inspect(pkg, func(n ast.Node) bool {
      switch x := n.(type) {
      case *ast.TypeSpec:

        result, _ := ParseStruct(x)
        fmt.Println(result.Name)

        for k, f := range result.Fields {
          // fmt.Println("type:  ", f.Type, " tag ", f.Tag)
          fmt.Println("  ", k, f)
        }
        // fmt.Println(result.Comments)
        fmt.Println(parseRequestComments(result.Comments))

      }

      return true
    })
  }

}
