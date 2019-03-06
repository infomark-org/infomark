package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"
)

type RequestField struct {
	Type    string
	Name    string
	Example string
	Childs  []RequestField
}

type Struct struct {
	Name        string
	Description string
	Fields      map[string]*Field
	Comments    *ast.CommentGroup
}

type Field struct {
	Tag  *Tag
	Type ast.Expr
}

func IdentString(str string, ch int) string {
	preSpace := strings.Repeat(" ", ch)
	lines := strings.Split(str, "\n")
	for k, _ := range lines {
		lines[k] = preSpace + lines[k]
	}

	return strings.Join(lines, "\n")
}

func GetRequestSchema(group *ast.CommentGroup) (string, error) {
	if group == nil {
		return "", errors.New("no comments")
	}
	txt := ""
	if group != nil {
		if len(group.List) == 1 {
			return "", errors.New("empty comments")
		}
		for k, el := range group.List {
			if el != nil {

				if k == 0 {
					if !strings.Contains(el.Text, "request payload") {
						return "", errors.New("no request schema")
					}
				} else {

					txt = txt + fmt.Sprintf("%s\n", el.Text[3:])
				}
			}
		}
	}
	return txt, nil
}

func GetResponseSchema(group *ast.CommentGroup) (string, error) {
	if group == nil {
		return "", errors.New("no comments")
	}
	txt := ""
	if group != nil {
		if len(group.List) == 1 {
			return "", errors.New("empty comments")
		}
		for k, el := range group.List {
			if el != nil {

				if k == 0 {
					if !strings.Contains(el.Text, "response payload") {
						return "", errors.New("no response schema")
					}
				} else {

					txt = txt + fmt.Sprintf("%s\n", el.Text[3:])
				}
			}
		}
	}
	return txt, nil
}

func ParseStruct(node ast.Node) (*Struct, error) {

	result := &Struct{}

	result.Name = node.(*ast.TypeSpec).Name.Name
	result.Comments = node.(*ast.TypeSpec).Doc

	return result, nil
}

func GetRequestStructs(pkg map[string]*ast.Package) map[string]*Struct {

	result := make(map[string]*Struct)

	for _, pkg := range pkg {
		ast.Inspect(pkg, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:

				s, err := ParseStruct(x)
				if err == nil {
					if strings.HasSuffix(s.Name, "Request") {
						result[s.Name] = s
					}
				}

			}

			return true
		})
	}

	return result
}

func GetResponseStructs(pkg map[string]*ast.Package) map[string]*Struct {

	result := make(map[string]*Struct)

	for _, pkg := range pkg {
		ast.Inspect(pkg, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:

				s, err := ParseStruct(x)
				if err == nil {
					if strings.HasSuffix(s.Name, "Response") {
						result[s.Name] = s
					}
				}

			}

			return true
		})
	}

	return result
}
