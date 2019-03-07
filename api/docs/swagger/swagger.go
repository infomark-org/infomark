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
  "strings"
)

// func SwaggerStructExample(structDescr *Struct, depth int) {
//   pre := strings.Repeat(" ", depth)
//   for _, fieldDescr := range structDescr.Fields {

//     if len(fieldDescr.Childs) > 0 {
//       for _, ch := range fieldDescr.Childs {
//         SwaggerStructExample(ch, depth+4)
//       }
//     } else {
//       fmt.Printf("%s  %s:\n", pre, fieldDescr.Tag.Name)

//     }
//   }
// }

func SwaggerStructs(structDescr *Struct, depth int) {
  pre := strings.Repeat(" ", depth)

  if len(structDescr.Fields) > 0 {
    fmt.Printf("%stype: object\n", pre)
    fmt.Printf("%sproperties:\n", pre)
  }

  requiredFields := []string{}
  examples := make(map[string]string)

  for _, fieldDescr := range structDescr.Fields {
    fmt.Printf("%s  %s:\n", pre, fieldDescr.Tag.Name)

    if len(fieldDescr.Childs) > 0 {
      for _, ch := range fieldDescr.Childs {
        SwaggerStructs(ch, depth+4)
      }

    } else {

      switch x := fieldDescr.Type.(type) {
      case *ast.Ident:
        switch x.Name {
        case "string":
          fmt.Printf("%s    type: %v\n", pre, fieldDescr.Type)
          if strings.Contains(fieldDescr.Tag.Name, "email") {
            fmt.Printf("%s    format: email\n", pre)
          }
        case "int":
          fmt.Printf("%s    type: integer\n", pre)
        case "bool":
          fmt.Printf("%s    type: boolean\n", pre)
        case "int64":
          fmt.Printf("%s    type: integer\n", pre)
          fmt.Printf("%s    format: %v\n", pre, fieldDescr.Type)
        case "float32", "float64":
          fmt.Printf("%s    type: number\n", pre)
          fmt.Printf("%s    format: %v\n", pre, fieldDescr.Type)
        default:

        }

        if fieldDescr.Tag.Example != "" {
          examples[fieldDescr.Tag.Name] = fieldDescr.Tag.Example
        }

      case *ast.SelectorExpr:
        if x.X.(*ast.Ident).Name == "null" && x.Sel.Name == "String" {
          fmt.Printf("%s    type: string\n", pre)
          fieldDescr.Tag.Required = false
        }

        if x.X.(*ast.Ident).Name == "time" && x.Sel.Name == "Time" {
          fmt.Printf("%s    type: string\n", pre)
          fmt.Printf("%s    format: date-time\n", pre)
          fieldDescr.Tag.Required = true

          examples[fieldDescr.Tag.Name] = "'2019-07-30T23:59:59Z'"
        }
      default:
        // spew.Dump(fieldDescr)
        // panic(fieldDescr)
        // panic(x)
      }

    }

    if fieldDescr.Tag.Required {
      requiredFields = append(requiredFields, fieldDescr.Tag.Name)

    }
  }

  fmt.Printf("%srequired:\n", pre)
  for _, f := range requiredFields {
    fmt.Printf("%s  - %s\n", pre, f)
  }

  if len(examples) > 0 {
    fmt.Printf("%sexample:\n", pre)
    for k, f := range examples {
      fmt.Printf("%s  %s: %s\n", pre, k, f)
    }

  }

}

func SwaggerStructsWithSuffix(pkgs map[string]*ast.Package, suffix string, depth int) {
  // gather structs
  for _, pkg := range pkgs {
    ast.Inspect(pkg, func(n ast.Node) bool {
      switch x := n.(type) {
      case *ast.TypeSpec:
        result, ok := x.Type.(*ast.StructType)
        if ok {
          name := n.(*ast.TypeSpec).Name.Name
          if strings.HasSuffix(name, suffix) {
            structDescr, err := ParseStruct(result, name, 0)
            if err == nil {
              pre := strings.Repeat(" ", depth)
              fmt.Printf("%s%s:\n", pre, name)

              SwaggerStructs(structDescr, depth+2)
            }
          }
        }
      }
      return true
    })
  }
}

func SwaggerResponsesWithSuffix(pkgs map[string]*ast.Package, suffix string, depth int) {
  pre := strings.Repeat(" ", depth)
  // gather structs
  for _, pkg := range pkgs {
    ast.Inspect(pkg, func(n ast.Node) bool {
      switch x := n.(type) {
      case *ast.TypeSpec:
        _, ok := x.Type.(*ast.StructType)
        if ok {
          name := n.(*ast.TypeSpec).Name.Name
          if strings.HasSuffix(name, suffix) {
            fmt.Printf("%s%s:\n", pre, name)
            fmt.Printf("%s  description: done\n", pre)
            fmt.Printf("%s  content:\n", pre)
            fmt.Printf("%s    application/json:\n", pre)
            fmt.Printf("%s      schema:\n", pre)
            fmt.Printf("%s        $ref: \"#/components/schemas/%s\"\n", pre, name)
          }
        }
      }
      return true
    })
  }
}
