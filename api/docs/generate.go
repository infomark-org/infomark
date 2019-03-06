package main

import (
    "fmt"
    "go/parser"
    "go/token"

    myparser "github.com/cgtuebingen/infomark-backend/api/docs/parser"
)

// Main docs
func main() {

    fset := token.NewFileSet() // positions are relative to fset

    // pkg, err := parser.ParseDir(fset, "./parser/fixture", nil, parser.ParseComments)
    pkg, err := parser.ParseDir(fset, "../app", nil, parser.ParseComments)
    if err != nil {
        panic(err)
    }

    requests := myparser.GetRequestStructs(pkg)
    responses := myparser.GetResponseStructs(pkg)
    endpoints := myparser.GetEndpoints(pkg)

    fmt.Printf("components:\n")
    fmt.Printf("  securitySchemes:\n")
    fmt.Printf("      bearerAuth:\n")
    fmt.Printf("        type: http\n")
    fmt.Printf("        scheme: bearer\n")
    fmt.Printf("        bearerFormat: JWT\n")
    fmt.Printf("      cookieAuth:\n")
    fmt.Printf("        type: apiKey\n")
    fmt.Printf("        in: cookie\n")
    fmt.Printf("        name: SESSIONID\n")
    fmt.Printf("  schemas:\n")
    for _, request := range requests {
        if request.Comments != nil {
            fmt.Printf("    %s:\n", request.Name)
            schema, err := myparser.GetRequestSchema(request.Comments)
            if err == nil {
                fmt.Println(myparser.IdentString(schema, 6))
            }
        }
    }
    fmt.Printf("  responses:\n")
    fmt.Printf("    NoContent:\n")
    fmt.Printf("      description: Update was successful.\n")
    fmt.Printf("    BadRequest:\n")
    fmt.Printf("      description: The request is in a wrong format or contains missing fields.\n")
    fmt.Printf("      content:\n")
    fmt.Printf("        application/json:\n")
    fmt.Printf("          schema:\n")
    fmt.Printf("            $ref: \"#/components/schemas/Error\"\n")
    fmt.Printf("    Unauthenticated:\n")
    fmt.Printf("      description: User is not logged in.\n")
    fmt.Printf("      content:\n")
    fmt.Printf("        application/json:\n")
    fmt.Printf("          schema:\n")
    fmt.Printf("            $ref: \"#/components/schemas/Error\"\n")
    fmt.Printf("    Unauthorized:\n")
    fmt.Printf("      description: User is logged in but has not the permission to perform the request.\n")
    fmt.Printf("      content:\n")
    fmt.Printf("        application/json:\n")
    fmt.Printf("          schema:\n")
    fmt.Printf("            $ref: \"#/components/schemas/Error\"\n")

    for _, response := range responses {
        if response.Comments != nil {
            fmt.Printf("    %s:\n", response.Name)
            schema, err := myparser.GetResponseSchema(response.Comments)
            if err == nil {
                fmt.Println(myparser.IdentString(schema, 6))
            }

        }
    }

    fmt.Printf("paths:\n")
    for url, _ := range endpoints {

        fmt.Printf("  %s:\n", url)
        for _, action := range endpoints[url] {

            fmt.Printf("    %s:\n", action.Details.Method)
            fmt.Printf("      summary: %s\n", action.Details.Summary)
            fmt.Printf("      description: >\n      %s\n", action.Details.Description)
            if action.Details.Request != "" {
                fmt.Printf("      requestBody:\n")
                fmt.Printf("        required: true\n")
                fmt.Printf("        content:\n")
                fmt.Printf("          application/json:\n")
                fmt.Printf("            schema:\n")
                fmt.Printf("              $ref: \"#/components/schemas/%s\"\n", action.Details.Request)
            }
            fmt.Printf("      responses:\n")
            for _, r := range action.Details.Responses {
                fmt.Printf("        \"%v\":\n", r.Code)
                fmt.Printf("          $ref: \"#/components/responses/%s\"\n", r.Text)
            }
        }

        // fmt.Println("  Method      ", endpoint.Details.Method)
        // fmt.Println("  Section     ", endpoint.Details.Section)
        // fmt.Println("  Request     ", endpoint.Details.Request)
        // fmt.Println("  Description ", endpoint.Details.Description)
    }

}
