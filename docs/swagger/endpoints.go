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
	"go/token"
	"strconv"
	"strings"
)

type Response struct {
	Code int
	Text string
}

func ParseResponse(source string) (*Response, error) {
	tmp := strings.Split(source, ",")
	stringCode := strings.TrimSpace(tmp[0])
	stringText := strings.TrimSpace(tmp[1])
	code, err := strconv.Atoi(stringCode)
	if err != nil {
		return nil, err
	}

	return &Response{Code: code, Text: stringText}, nil

}

func ParseParameter(source string) (*Parameter, error) {
	tmp := strings.Split(source, ",")
	if len(tmp) != 2 {
		return nil, fmt.Errorf("error in \"%s\"", source)
	}
	stringName := strings.TrimSpace(tmp[0])
	stringType := strings.TrimSpace(tmp[1])

	return &Parameter{Name: stringName, Type: stringType}, nil
}

type Endpoint struct {
	Node     ast.Node
	Info     *ast.FuncDecl
	Comments *ast.CommentGroup
	Details  EndpointDetails
	Position token.Position
}

type Parameter struct {
	Name string
	Type string
}

type EndpointDetails struct {
	URL         string
	Method      string
	URLParams   []*Parameter
	QueryParams []*Parameter
	Request     string
	Tags        []string
	Responses   []*Response
	Description string
	Summary     string
}

func NewEndpoint(node ast.Node) *Endpoint {
	return &Endpoint{
		Node:     node,
		Info:     node.(*ast.FuncDecl),
		Comments: node.(*ast.FuncDecl).Doc,
		Details:  parseComments(node.(*ast.FuncDecl).Doc),
	}
}

func parseComments(group *ast.CommentGroup) EndpointDetails {
	descrp := EndpointDetails{}
	descrStart := -1
	for k, el := range group.List {
		if el != nil {
			// fmt.Println(el.Text)
			if strings.Contains(el.Text, "METHOD:") {
				tmp := strings.Split(el.Text, ":")
				descrp.Method = strings.TrimSpace(tmp[1])
			}
			if strings.Contains(el.Text, "URL:") {
				tmp := strings.Split(el.Text, ":")
				descrp.URL = strings.TrimSpace(tmp[1])
			}
			if strings.Contains(el.Text, "REQUEST:") {
				tmp := strings.Split(el.Text, ":")
				descrp.Request = strings.TrimSpace(tmp[1])
			}
			if strings.Contains(el.Text, "TAG:") {
				tmp := strings.Split(el.Text, ":")

				descrp.Tags = append(descrp.Tags, strings.TrimSpace(tmp[1]))
			}
			if strings.Contains(el.Text, "SUMMARY:") {
				tmp := strings.Split(el.Text, ":")
				descrp.Summary = strings.TrimSpace(tmp[1])
			}
			if strings.Contains(el.Text, "RESPONSE:") {
				tmp := strings.Split(el.Text, ":")
				resp, err := ParseResponse(tmp[1])
				if err == nil {
					// descrp.Responses = append(descrp.Responses, strings.TrimSpace(tmp[1]))
					descrp.Responses = append(descrp.Responses, resp)
				}
			}
			if strings.Contains(el.Text, "URLPARAM:") {
				tmp := strings.Split(el.Text, ":")
				resp, err := ParseParameter(tmp[1])
				if err == nil {
					descrp.URLParams = append(descrp.URLParams, resp)
				}
			}
			if strings.Contains(el.Text, "QUERYPARAM:") {
				tmp := strings.Split(el.Text, ":")
				resp, err := ParseParameter(tmp[1])
				if err == nil {
					descrp.QueryParams = append(descrp.QueryParams, resp)
				}
			}
			if strings.Contains(el.Text, "DESCRIPTION:") {
				descrStart = k + 1
				break
			}
		}
	}

	if descrStart > -1 {
		text := ""
		for k, el := range group.List {
			if k >= descrStart {
				text = text + el.Text[2:]
			}
		}
		descrp.Description = text
	}
	return descrp
}

func isPublicEndpoint(group *ast.CommentGroup) (string, bool) {
	if group != nil {
		if group.List != nil {
			for _, el := range group.List {
				if el != nil {
					// fmt.Println(el.Text)
					if strings.Contains(el.Text, "is public endpoint for") {
						return el.Text, true
					}
				}
			}
		}
	}
	return "", false
}

// Main docs
func GetEndpoints(pkg map[string]*ast.Package, fset *token.FileSet) map[string][]*Endpoint {

	result := make(map[string][]*Endpoint)
	// fmt.Println(fset)
	// var endpoints []*Endpoint

	// gather endpoints
	for _, pkg := range pkg {
		ast.Inspect(pkg, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if str, is := isPublicEndpoint(x.Doc); is {
					// commentNames := strings.Split(str[2:], " ")
					// commentName := strings.TrimSpace(commentNames[0])
					functionName := strings.TrimSpace(n.(*ast.FuncDecl).Name.Name)
					// fmt.Println("---------------", str)
					// fmt.Println("--", functionName, fset.Position(n.Pos()))
					// fmt.Println("--")

					if !strings.Contains(str, functionName) {
						msg := fmt.Sprintf("\"%s\" does not contains \"%s\" in %v", str, functionName, fset.Position(n.Pos()))
						panic(msg)
					}

					ep := NewEndpoint(x)
					ep.Position = fset.Position(n.Pos())
					result[ep.Details.URL] = append(result[ep.Details.URL], ep)
				}
			}

			return true
		})
	}

	return result
}
