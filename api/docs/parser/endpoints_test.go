package parser

import (
    "fmt"
    "go/parser"
    "go/token"
    "testing"
)

// Main docs
func TestEndpoint(t *testing.T) {

    fset := token.NewFileSet() // positions are relative to fset

    pkg, err := parser.ParseDir(fset, "./fixture", nil, parser.ParseComments)
    if err != nil {
        panic(err)
    }

    endpoints := GetEndpoints(pkg)

    fmt.Println("\n\n\n\n\n  ")

    // for _, endpoint := range endpoints {
    //     fmt.Println("Endpoint      ", endpoint.Details.URL)
    //     fmt.Println("  Method      ", endpoint.Details.Method)
    //     fmt.Println("  Section     ", endpoint.Details.Section)
    //     fmt.Println("  Request     ", endpoint.Details.Request)
    //     for _, r := range endpoint.Details.Responses {
    //         fmt.Println("  Response    ", r.Code, "(", r.Text, ")")

    //     }
    //     fmt.Println("  Description ", endpoint.Details.Description)
    // }
    //
    for url, _ := range endpoints {

        fmt.Printf("  %s:\n", url)
        for _, action := range endpoints[url] {

            fmt.Printf("    %s:\n", action.Details.Method)
            fmt.Printf("      summary: %s\n", action.Details.Summary)
            fmt.Printf("      description: >\n      %s\n", action.Details.Description)
            fmt.Printf("      responses:\n")
            for _, r := range action.Details.Responses {
                fmt.Printf("        \"%v\":\n", r.Code)
                fmt.Printf("          $ref: \"#/components/responses/%s\"\n", r.Text)
            }
            if action.Details.Request != "" {
                fmt.Printf("      requestBody:\n")
                fmt.Printf("        required: true\n")
                fmt.Printf("        content:\n")
                fmt.Printf("          application/json:\n")
                fmt.Printf("            schema:\n")
                fmt.Printf("              $ref: \"#/components/schemas/%s\"\n", action.Details.Request)

            }
        }

        // fmt.Println("  Method      ", endpoint.Details.Method)
        // fmt.Println("  Section     ", endpoint.Details.Section)
        // fmt.Println("  Request     ", endpoint.Details.Request)
        // fmt.Println("  Description ", endpoint.Details.Description)
    }
    fmt.Println("\n\n\n\n\n  ")
}
