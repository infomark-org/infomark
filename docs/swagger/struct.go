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
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type Struct struct {
	Name        string
	Description string
	Fields      []*Field
	Comments    *ast.CommentGroup
	Tag         *Tag
	Position    token.Position
}

type Field struct {
	Tag    *Tag
	Type   ast.Expr
	Childs []*Struct
	Depth  int
}

func IdentString(str string, ch int) string {
	preSpace := strings.Repeat(" ", ch)
	lines := strings.Split(str, "\n")
	for k, _ := range lines {
		lines[k] = preSpace + lines[k]
	}

	return strings.Join(lines, "\n")
}

func ParseField(fset *token.FileSet, n ast.Node, field *ast.Field, depth int) (*Field, error) {
	result := &Field{}
	result.Type = field.Type
	result.Depth = depth

	t, err := parseTag(string(field.Tag.Value))
	if err == nil {

		if t.Name == "-" {
			return nil, errors.New("Skip -")
		}
		result.Tag = t
	} else {
		return nil, err
	}

	if len(field.Names) > 0 {

		switch x := field.Type.(type) {
		case *ast.StructType:
			child, err := ParseStruct(fset, n, x, field.Names[0].Name, depth+1)
			if err == nil {
				result.Childs = append(result.Childs, child)

			}
		case *ast.StarExpr:
			switch y := x.X.(type) {
			case *ast.StructType:
				child, err := ParseStruct(fset, n, y, field.Names[0].Name, depth+1)
				if err == nil {
					result.Childs = append(result.Childs, child)

				}
			}

		}

	}

	return result, nil
}

func ParseStruct(fset *token.FileSet, n ast.Node, structDecl *ast.StructType, name string, depth int) (*Struct, error) {
	fields := structDecl.Fields.List

	result := &Struct{Name: name}
	result.Position = fset.Position(n.Pos())

	for _, field := range fields {
		if field.Tag != nil {
			t, err := parseTag(string(field.Tag.Value))
			if err == nil {

				result.Tag = t
			} else {
				spew.Dump(structDecl)
				fmt.Println(string(field.Tag.Value))
				panic(err)
			}

			f, err := ParseField(fset, n, field, depth)
			if err == nil {
				result.Fields = append(result.Fields, f)
			}
		}
	}

	return result, nil
}
