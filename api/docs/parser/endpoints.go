package parser

import (
    "go/ast"
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

type Endpoint struct {
    Node     ast.Node
    Info     *ast.FuncDecl
    Comments *ast.CommentGroup
    Details  EndpointDetails
}

type EndpointDetails struct {
    URL         string
    Method      string
    Request     string
    Section     string
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
            if strings.Contains(el.Text, "METHOD") {
                tmp := strings.Split(el.Text, ":")
                descrp.Method = strings.TrimSpace(tmp[1])
            }
            if strings.Contains(el.Text, "URL") {
                tmp := strings.Split(el.Text, ":")
                descrp.URL = strings.TrimSpace(tmp[1])
            }
            if strings.Contains(el.Text, "REQUEST") {
                tmp := strings.Split(el.Text, ":")
                descrp.Request = strings.TrimSpace(tmp[1])
            }
            if strings.Contains(el.Text, "SECTION") {
                tmp := strings.Split(el.Text, ":")
                descrp.Section = strings.TrimSpace(tmp[1])
            }
            if strings.Contains(el.Text, "SUMMARY") {
                tmp := strings.Split(el.Text, ":")
                descrp.Summary = strings.TrimSpace(tmp[1])
            }
            if strings.Contains(el.Text, "RESPONSE") {
                tmp := strings.Split(el.Text, ":")
                resp, err := ParseResponse(tmp[1])
                if err == nil {
                    // descrp.Responses = append(descrp.Responses, strings.TrimSpace(tmp[1]))
                    descrp.Responses = append(descrp.Responses, resp)

                }
            }
            if strings.Contains(el.Text, "DESCRIPTION") {
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

func isPublicEndpoint(group *ast.CommentGroup) bool {
    if group != nil {
        if group.List != nil {
            for _, el := range group.List {
                if el != nil {
                    // fmt.Println(el.Text)
                    if strings.Contains(el.Text, "is public endpoint for") {
                        return true
                    }
                }
            }
        }
    }
    return false
}

// Main docs
func GetEndpoints(pkg map[string]*ast.Package) map[string][]*Endpoint {

    result := make(map[string][]*Endpoint)

    // var endpoints []*Endpoint

    // gather endpoints
    for _, pkg := range pkg {
        ast.Inspect(pkg, func(n ast.Node) bool {
            switch x := n.(type) {
            case *ast.FuncDecl:
                if isPublicEndpoint(x.Doc) {
                    ep := NewEndpoint(x)
                    result[ep.Details.URL] = append(result[ep.Details.URL], ep)
                }
            }

            return true
        })
    }

    return result
}
