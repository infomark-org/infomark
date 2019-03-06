package main

import (
    "errors"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "strconv"
    "strings"
)

type Tag struct {
    // Key is the tag key, such as json, xml, etc..
    // i.e: `json:"foo,omitempty". Here key is: "json"
    Key string

    // Name is a part of the value
    // i.e: `json:"foo,omitempty". Here name is: "foo"
    Name string

    // Options is a part of the value. It contains a slice of tag options i.e:
    // `json:"foo,omitempty". Here options is: ["omitempty"]
    Options []string
}

var (
    errTagSyntax      = errors.New("bad syntax for struct tag pair")
    errTagKeySyntax   = errors.New("bad syntax for struct tag key")
    errTagValueSyntax = errors.New("bad syntax for struct tag value")

    errKeyNotSet      = errors.New("tag key does not exist")
    errTagNotExist    = errors.New("tag does not exist")
    errTagKeyMismatch = errors.New("mismatch between key and tag.key")
)

func parseTag(tag string) (*RequestField, error) {
    var tags []*Tag

    tag = tag[1 : len(tag)-1]

    // NOTE(arslan) following code is from reflect and vet package with some
    // modifications to collect all necessary information and extend it with
    // usable methods
    for tag != "" {
        // fmt.Println("parse:", tag)
        if len(tag) < 3 {
            return nil, nil
        }
        // Skip leading space.
        i := 0
        for i < len(tag) && tag[i] == ' ' {
            i++
        }
        tag = tag[i:]
        if tag == "" {
            return nil, nil
        }

        // Scan to colon. A space, a quote or a control character is a syntax
        // error. Strictly speaking, control chars include the range [0x7f,
        // 0x9f], not just [0x00, 0x1f], but in practice, we ignore the
        // multi-byte control characters as it is simpler to inspect the tag's
        // bytes than the tag's runes.
        i = 0
        for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
            i++
        }

        if i == 0 {
            return nil, errTagKeySyntax
        }
        if i+1 >= len(tag) || tag[i] != ':' {
            return nil, errTagSyntax
        }
        if tag[i+1] != '"' {
            return nil, errTagValueSyntax
        }

        key := string(tag[:i])
        tag = tag[i+1:]

        // Scan quoted string to find value.
        i = 1
        for i < len(tag) && tag[i] != '"' {
            if tag[i] == '\\' {
                i++
            }
            i++
        }
        if i >= len(tag) {
            return nil, errTagValueSyntax
        }

        qvalue := string(tag[:i+1])
        tag = tag[i+1:]

        value, err := strconv.Unquote(qvalue)
        if err != nil {
            return nil, errTagValueSyntax
        }

        res := strings.Split(value, ",")
        name := res[0]
        options := res[1:]
        if len(options) == 0 {
            options = nil
        }
        // fmt.Println("got ", key, name)
        tags = append(tags, &Tag{
            Key:     key,
            Name:    name,
            Options: options,
        })
    }

    field := &RequestField{}
    for _, tag := range tags {

        if tag.Key == "json" {
            field.Name = tag.Name
        }

        if tag.Key == "example" {
            field.Example = tag.Name
        }

        if tag.Key == "required" {
            field.Required = true
        }
    }

    return field, nil
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
    Description string
}

func NewEndpoint(node ast.Node) Endpoint {
    return Endpoint{
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
                descrp.Method = tmp[1]
            }
            if strings.Contains(el.Text, "URL") {
                tmp := strings.Split(el.Text, ":")
                descrp.URL = tmp[1]
            }
            if strings.Contains(el.Text, "REQUEST") {
                tmp := strings.Split(el.Text, ":")
                descrp.Request = tmp[1]
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

type Request struct {
    Node     ast.Node
    Info     *ast.TypeSpec
    Comments *ast.CommentGroup
    Details  RequestDetails
}

type RequestDetails struct {
    Description string
    Fields      []*RequestField
}

type RequestField struct {
    Name     string
    Example  string
    Required bool
}

func NewRequest(node ast.Node) Request {
    structDecl := node.(*ast.TypeSpec).Type.(*ast.StructType)
    fields := structDecl.Fields.List
    // fmt.Println(node)
    // fmt.Println(parseRequestDescription(node.(*ast.TypeSpec).Doc))

    rFields := []*RequestField{}
    for _, field := range fields {
        // typeExpr := field.Type

        // start := typeExpr.Pos() - 1
        // end := typeExpr.End() - 1

        // // grab it in source
        // typeInSource := src[start:end]

        // fmt.Println(typeInSource)
        // fmt.Println(field)
        if len(field.Names) > 0 {
            // fmt.Println(field.Names[0])
            // fmt.Println(field.Tag.Value)
            // tags, err := structtag.Parse(string(field.Tag.Value))
            field, err := parseTag(string(field.Tag.Value))

            if err != nil {
                panic(err)
            }

            // fmt.Println(field.Name, field.Example)
            rFields = append(rFields, field)

        }
    }

    r := Request{
        Node:     node,
        Info:     node.(*ast.TypeSpec),
        Comments: node.(*ast.TypeSpec).Doc,
        Details: RequestDetails{
            Description: parseRequestDescription(node.(*ast.TypeSpec).Doc),
            Fields:      rFields,
        },
    }

    // fmt.Println(r.Info.Name.Name, r)
    return r
}

func parseRequestDescription(group *ast.CommentGroup) string {
    txt := ""
    if group != nil {
        for _, el := range group.List {
            if el != nil {
                txt = txt + el.Text[2:]
            }
        }
    }
    return txt
}

func isRequestStruct(t *ast.TypeSpec) bool {
    if t.Name != nil {
        return strings.HasSuffix(t.Name.Name, "Request")

    }
    return false
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
func main() {

    var endpoints []Endpoint
    // var requests []Request

    requestMap := make(map[string]Request)

    fset := token.NewFileSet() // positions are relative to fset

    d, err := parser.ParseDir(fset, "../app", nil, parser.ParseComments)
    if err != nil {
        fmt.Println(err)
        return
    }

    // gather structs
    for _, pkg := range d {
        ast.Inspect(pkg, func(n ast.Node) bool {
            switch x := n.(type) {
            case *ast.TypeSpec:
                if isRequestStruct(x) {
                    // requests = append(requests, NewRequest(x))
                    r := NewRequest(x)
                    requestMap[strings.TrimSpace(r.Info.Name.Name)] = r
                }

            }

            return true
        })
    }

    // for k, el := range requestMap {
    //     fmt.Println(k, el)
    // }

    // gather endpoints
    for _, pkg := range d {
        ast.Inspect(pkg, func(n ast.Node) bool {
            switch x := n.(type) {
            case *ast.FuncDecl:
                if isPublicEndpoint(x.Doc) {
                    endpoints = append(endpoints, NewEndpoint(x))
                }
            }

            return true
        })
    }

    // for _, request := range requests {
    //     fmt.Println(request.Info.Name, request.Info.Doc, request.Comments, request.Details.Description)
    //     for _, f := range request.Details.Fields {
    //         fmt.Println("  ", f.Name, f.Example, f.Required)
    //     }
    // }

    for _, endpoint := range endpoints {
        fmt.Println("Endpoint      ", endpoint.Details.URL)
        fmt.Println("  Method      ", endpoint.Details.Method)
        fmt.Println("  Description ", endpoint.Details.Description)
        fmt.Println("  Request     ", endpoint.Details.Request)
        for _, f := range requestMap[strings.TrimSpace(endpoint.Details.Request)].Details.Fields {
            fmt.Println("   ", f.Name, f.Example, f.Required)
        }
        // fmt.Println(endpoint.Info.Name, endpoint.Info.Doc, endpoint.Comments, endpoint.Details)
        // fmt.Println(endpoint.Details.Request)
        // fmt.Println(requestMap[endpoint.Details.Request])
    }
    //
    // for k, pkg := range d {
    //     fmt.Println("package", k)
    //     p := doc.New(pkg, "./", 0)

    //     for _, t := range p.Types {
    //         fmt.Println("  type", t.Name)
    //         fmt.Println("    docs:", t.Doc)
    //     }
    // }

}
