package main

import (
    "fmt"
    "go/parser"
    "go/token"
    "strings"

    "github.com/cgtuebingen/infomark-backend/api/docs/swagger"
)

// Main docs
func main() {

    fset := token.NewFileSet() // positions are relative to fset

    pkgs, err := parser.ParseDir(fset, "../app", nil, parser.ParseComments)
    if err != nil {
        panic(err)
    }

    endpoints := swagger.GetEndpoints(pkgs)

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
    fmt.Printf("    Error:\n")
    fmt.Printf("      type: object\n")
    fmt.Printf("      properties:\n")
    fmt.Printf("        code:\n")
    fmt.Printf("          type: string\n")
    fmt.Printf("        message:\n")
    fmt.Printf("          type: string\n")
    fmt.Printf("      required:\n")
    fmt.Printf("        - code\n")
    fmt.Printf("        - message\n")

    swagger.SwaggerStructsWithSuffix(pkgs, "Request", 4)
    swagger.SwaggerStructsWithSuffix(pkgs, "Response", 4)
    fmt.Printf("  responses:\n")
    fmt.Printf("    pongResponse:\n")
    fmt.Printf("      description: Server is up and running\n")
    fmt.Printf("      content:\n")
    fmt.Printf("        text/plain:\n")
    fmt.Printf("          schema:\n")
    fmt.Printf("            type: string\n")
    fmt.Printf("            example: pong\n")
    fmt.Printf("    ZipFile:\n")
    fmt.Printf("      description: A file as a download.\n")
    fmt.Printf("      content:\n")
    fmt.Printf("        application/zip:\n")
    fmt.Printf("          schema:\n")
    fmt.Printf("            type: string\n")
    fmt.Printf("            format: binary\n")
    fmt.Printf("    ImageFile:\n")
    fmt.Printf("      description: A file as a download.\n")
    fmt.Printf("      content:\n")
    fmt.Printf("        image/jpeg:\n")
    fmt.Printf("          schema:\n")
    fmt.Printf("            type: string\n")
    fmt.Printf("            format: binary\n")
    fmt.Printf("    OK:\n")
    fmt.Printf("      description: Post successfully delivered.\n")
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

    // create all responses
    swagger.SwaggerResponsesWithSuffix(pkgs, "Response", 4)

    // create all list responses
    pre := strings.Repeat(" ", 4)
    for url, _ := range endpoints {
        for _, action := range endpoints[url] {
            for _, r := range action.Details.Responses {
                text := strings.TrimSpace(r.Text)
                if strings.HasSuffix(text, "List") {
                    fmt.Printf("%s%s:\n", pre, text)
                    fmt.Printf("%s  description: done\n", pre)
                    fmt.Printf("%s  content:\n", pre)
                    fmt.Printf("%s    application/json:\n", pre)
                    fmt.Printf("%s      schema:\n", pre)
                    fmt.Printf("%s        type: array\n", pre)
                    fmt.Printf("%s        items:\n", pre)
                    fmt.Printf("%s          $ref: \"#/components/schemas/%s\"\n", pre, text[:len(text)-4])
                }

            }
        }
    }

    fmt.Printf("paths:\n")
    for url, _ := range endpoints {

        fmt.Printf("  %s:\n", url)
        for _, action := range endpoints[url] {

            fmt.Printf("    %s:\n", action.Details.Method)
            fmt.Printf("      summary: %s\n", action.Details.Summary)
            if action.Details.Description != "" {
                fmt.Printf("      description: >\n      %s\n", action.Details.Description)
            }
            if len(action.Details.Tags) > 0 {
                fmt.Printf("      tags: \n\n")
                for _, tag := range action.Details.Tags {

                    fmt.Printf("        - %s\n", tag)
                }
            }
            if len(action.Details.URLParams)+len(action.Details.QueryParams) > 0 {
                fmt.Printf("      parameters:\n")
                for _, el := range action.Details.URLParams {
                    fmt.Printf("        - in: path\n")
                    fmt.Printf("          name: %s\n", el.Name)
                    fmt.Printf("          schema: \n")
                    fmt.Printf("             type: %s\n", el.Type)
                    fmt.Printf("          required: true \n")
                }

                for _, el := range action.Details.QueryParams {
                    fmt.Printf("        - in: query\n")
                    fmt.Printf("          name: %s\n", el.Name)
                    fmt.Printf("          schema: \n")
                    fmt.Printf("             type: %s\n", el.Type)
                    fmt.Printf("          required: false \n")
                }
            }

            if action.Details.Request != "" {
                fmt.Printf("      requestBody:\n")
                fmt.Printf("        required: true\n")

                switch action.Details.Request {
                case "zipfile":
                    fmt.Printf("        content:\n")
                    fmt.Printf("          multipart/form-data:\n")
                    fmt.Printf("            schema:\n")
                    fmt.Printf("              type: object\n")
                    fmt.Printf("              properties:\n")
                    fmt.Printf("                file_data:\n")
                    fmt.Printf("                  type: string\n")
                    fmt.Printf("                  format: binary\n")
                    fmt.Printf("            encoding:\n")
                    fmt.Printf("              file_data:\n")
                    fmt.Printf("                contentType: application/zip\n")
                case "imagefile":
                    fmt.Printf("        content:\n")
                    fmt.Printf("          multipart/form-data:\n")
                    fmt.Printf("            schema:\n")
                    fmt.Printf("              type: object\n")
                    fmt.Printf("              properties:\n")
                    fmt.Printf("                file_data:\n")
                    fmt.Printf("                  type: string\n")
                    fmt.Printf("                  format: binary\n")
                    fmt.Printf("            encoding:\n")
                    fmt.Printf("              file_data:\n")
                    fmt.Printf("                contentType: image/jpeg\n")
                default:
                    fmt.Printf("        content:\n")
                    fmt.Printf("          application/json:\n")
                    fmt.Printf("            schema:\n")
                    fmt.Printf("              $ref: \"#/components/schemas/%s\"\n", action.Details.Request)
                }

            }
            fmt.Printf("      responses:\n")
            for _, r := range action.Details.Responses {

                fmt.Printf("        \"%v\":\n", r.Code)
                fmt.Printf("          $ref: \"#/components/responses/%s\"\n", r.Text)

            }
        }

    }

}
