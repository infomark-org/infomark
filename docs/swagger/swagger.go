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

package swagger

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func SwaggerStructs(structDescr *Struct, depth int) string {
	source := ""
	pre := strings.Repeat(" ", depth)

	if len(structDescr.Fields) > 0 {
		source = source + fmt.Sprintf("%stype: object\n", pre)
		source = source + fmt.Sprintf("%sproperties:\n", pre)
	}

	requiredFields := []string{}
	examples := make(map[string]string)

	for _, fieldDescr := range structDescr.Fields {
		source = source + fmt.Sprintf("%s  %s:\n", pre, fieldDescr.Tag.Name)

		if len(fieldDescr.Childs) > 0 {
			for _, ch := range fieldDescr.Childs {
				source = source + SwaggerStructs(ch, depth+4)
			}

		} else {

			switch x := fieldDescr.Type.(type) {
			case *ast.Ident:
				switch x.Name {
				case "string":
					source = source + fmt.Sprintf("%s    type: %v\n", pre, fieldDescr.Type)
					if strings.Contains(fieldDescr.Tag.Name, "email") {
						source = source + fmt.Sprintf("%s    format: email\n", pre)
					}
					if strings.Contains(fieldDescr.Tag.Name, "password") {
						source = source + fmt.Sprintf("%s    format: password\n", pre)
					}
				case "int":
					source = source + fmt.Sprintf("%s    type: integer\n", pre)
				case "bool":
					source = source + fmt.Sprintf("%s    type: boolean\n", pre)
				case "int64":
					source = source + fmt.Sprintf("%s    type: integer\n", pre)
					source = source + fmt.Sprintf("%s    format: %v\n", pre, fieldDescr.Type)
				case "float32", "float64":
					source = source + fmt.Sprintf("%s    type: number\n", pre)
					source = source + fmt.Sprintf("%s    format: %v\n", pre, fieldDescr.Type)
				default:

				}

				if fieldDescr.Tag.Length != "" {
					source = source + fmt.Sprintf("%s    length: %s\n", pre, fieldDescr.Tag.Length)
				}
				if fieldDescr.Tag.MinValue != "" {
					source = source + fmt.Sprintf("%s    minimum: %s\n", pre, fieldDescr.Tag.MinValue)
				}
				if fieldDescr.Tag.MaxValue != "" {
					source = source + fmt.Sprintf("%s    maximum: %s\n", pre, fieldDescr.Tag.MaxValue)
				}

				if fieldDescr.Tag.Example == "" {
					if !strings.HasPrefix(structDescr.Name, "Err") {
						panic(fmt.Sprintf("field '%s' has no example in struct '%v'", fieldDescr.Tag.Name, structDescr.Name))
					}
					//
				} else {
					examples[fieldDescr.Tag.Name] = fieldDescr.Tag.Example
				}

			case *ast.SelectorExpr:
				if x.X.(*ast.Ident).Name == "null" && x.Sel.Name == "String" {
					source = source + fmt.Sprintf("%s    type: string\n", pre)
					fieldDescr.Tag.Required = false
				}

				if x.X.(*ast.Ident).Name == "time" && x.Sel.Name == "Time" {
					source = source + fmt.Sprintf("%s    type: string\n", pre)
					source = source + fmt.Sprintf("%s    format: date-time\n", pre)
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

	source = source + fmt.Sprintf("%srequired:\n", pre)
	for _, f := range requiredFields {
		source = source + fmt.Sprintf("%s  - %s\n", pre, f)
	}

	if len(examples) > 0 {
		source = source + fmt.Sprintf("%sexample:\n", pre)
		for k, f := range examples {
			source = source + fmt.Sprintf("%s  %s: %s\n", pre, k, f)
		}

	}

	return source

}

func SwaggerStructsWithSuffix(fset *token.FileSet, pkgs map[string]*ast.Package, suffix string, depth int) string {
	source := ""
	// gather structs
	for _, pkg := range pkgs {
		ast.Inspect(pkg, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				result, ok := x.Type.(*ast.StructType)
				if ok {
					name := n.(*ast.TypeSpec).Name.Name
					if strings.HasSuffix(name, suffix) {
						structDescr, err := ParseStruct(fset, n, result, name, 0)
						if err == nil {
							pre := strings.Repeat(" ", depth)
							source = source + fmt.Sprintf("%s# implementation in %v:\n", pre, structDescr.Position)
							source = source + fmt.Sprintf("%s%s:\n", pre, name)

							source = source + SwaggerStructs(structDescr, depth+2)
						}
					}
				}
			}
			return true
		})
	}
	return source
}

func SwaggerResponsesWithSuffix(fset *token.FileSet, pkgs map[string]*ast.Package, suffix string, depth int) string {
	source := ""
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
						source = source + fmt.Sprintf("%s# implementation in %v:\n", pre, fset.Position(n.Pos()))
						source = source + fmt.Sprintf("%s%s:\n", pre, name)
						source = source + fmt.Sprintf("%s  description: done\n", pre)
						source = source + fmt.Sprintf("%s  content:\n", pre)
						source = source + fmt.Sprintf("%s    application/json:\n", pre)
						source = source + fmt.Sprintf("%s      schema:\n", pre)
						source = source + fmt.Sprintf("%s        $ref: \"#/components/schemas/%s\"\n", pre, name)
					}
				}
			}
			return true
		})
	}

	return source
}
