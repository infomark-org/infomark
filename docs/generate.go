//go:build !codeanalysis
// +build !codeanalysis

package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/infomark-org/infomark/api/app"
	"github.com/infomark-org/infomark/configuration"
	"github.com/infomark-org/infomark/docs/swagger"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Routes struct {
	Method string
	Path   string
}

func EmptyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(""))
	})
}

func GetAllRoutes() []*Routes {
	db, _ := sqlx.Open("sqlite3", ":memory:")
	r, _ := app.New(db, EmptyHandler(), false)

	routes := []*Routes{}

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		route = strings.Replace(route, "/*/", "/", -1)
		route = strings.Replace(route, "/*/", "/", -1)
		route = strings.Replace(route, "/*/", "/", -1)
		route = strings.Replace(route, "/*", "/", -1)
		// f.WriteString(fmt.Sprintf("%s %s\n", method, route))

		if len(route) > 1 && strings.HasPrefix(route, "/api/v1") {
			route = strings.TrimSuffix(route, "/")
			routes = append(routes, &Routes{Method: strings.ToLower(method), Path: route[7:]})
		}
		return nil
	}

	if err := chi.Walk(r, walkFunc); err != nil {
		panic(fmt.Sprintf("Logging err: %s\n", err.Error()))
	}

	return routes
}

// Main docs
func main() {

	fset := token.NewFileSet() // positions are relative to fset

	configuration.MustFindAndReadConfiguration()

	pkgs, err := parser.ParseDir(fset, "./api/app/", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	endpoints := swagger.GetEndpoints(pkgs, fset)

	// verify we have all api-routes in the swagger description
	routes := GetAllRoutes()
	for _, route := range routes {
		found := false
		fmt.Println("Verify documentation for", route.Path)
		for url := range endpoints {
			if route.Path == url {

				for _, action := range endpoints[url] {
					// fmt.Println(action.Details.URL, action.Details.Method)
					if action.Details.Method == route.Method {
						found = true
						break
					}
				}

				if !found {
					candidates := []string{}
					for _, action := range endpoints[url] {
						candidates = append(candidates, fmt.Sprintf("%v", action.Position))
					}
					fmt.Printf("  - The documentation for the route '%s:%s' has mistakes and cannot be found.\n", route.Method, route.Path)
					fmt.Println("  - Is the url and method in the comment correct")

					panic(fmt.Sprintf("found '%s' but not the correct method want '%s' for route %v \n candiates are %v", url, route.Method, route, candidates))
				}
			}
			if found {
				break
			}
		}

		if !found {
			panic(fmt.Sprintf("did not found %v", route.Path))
		}
	}

	f, err := os.Create("./api.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// w := bufio.NewWriter(f)

	f.WriteString("# this file has been automatically created. Please do not edit it here\n")
	f.WriteString("openapi: 3.0.0\n")
	f.WriteString("info:\n")
	f.WriteString("  title: \"InfoMark\"\n")
	f.WriteString("  version: \"0.0.1\"\n")
	f.WriteString("  description: >\n")
	f.WriteString("    A CI based course framework. All enums should be send as strings and returned as strings.\n")
	f.WriteString("    Everything\n")
	f.WriteString("  contact:\n")
	f.WriteString("    - name: Mark Boss\n")
	f.WriteString("      email: mark.boss@uni-tuebingen.de\n")
	f.WriteString("      url: https://uni-tuebingen.de\n")
	f.WriteString("    - name: Patrick Wieschollek\n")
	f.WriteString("      email: Patrick.Wieschollek@uni-tuebingen.de\n")
	f.WriteString("      url: https://uni-tuebingen.de\n")
	f.WriteString("servers:\n")
	f.WriteString("  - url: http://localhost:2020/api/v1\n")
	f.WriteString("security:\n")
	f.WriteString("  - bearerAuth: []\n")
	f.WriteString("  - cookieAuth: []\n")
	f.WriteString("tags:\n")
	f.WriteString("  - name: common\n")
	f.WriteString("    description: common request\n")
	f.WriteString("  - name: auth\n")
	f.WriteString("    description: authenticated related requests\n")
	f.WriteString("  - name: account\n")
	f.WriteString("    description: account related requests\n")
	f.WriteString("  - name: email\n")
	f.WriteString("    description: Email related requests\n")
	f.WriteString("  - name: users\n")
	f.WriteString("    description: User related requests\n")
	f.WriteString("  - name: courses\n")
	f.WriteString("    description: Course related requests\n")
	f.WriteString("  - name: sheets\n")
	f.WriteString("    description: Exercise sheets related requests\n")
	f.WriteString("  - name: tasks\n")
	f.WriteString("    description: Exercise tasks related requests\n")
	f.WriteString("  - name: submissions\n")
	f.WriteString("    description: Submissions related requests\n")
	f.WriteString("  - name: grades\n")
	f.WriteString("    description: Gradings related requests\n")
	f.WriteString("  - name: groups\n")
	f.WriteString("    description: Exercise groups related requests\n")
	f.WriteString("  - name: enrollments\n")
	f.WriteString("    description: Enrollments related requests\n")
	f.WriteString("  - name: materials\n")
	f.WriteString("    description: Exercise material related requests\n")
	f.WriteString("  - name: internal\n")
	f.WriteString("    description: Endpoints for internal usage only\n")

	f.WriteString("components:\n")
	f.WriteString("  securitySchemes:\n")
	f.WriteString("      bearerAuth:\n")
	f.WriteString("        type: http\n")
	f.WriteString("        scheme: bearer\n")
	f.WriteString("        bearerFormat: JWT\n")
	f.WriteString("      cookieAuth:\n")
	f.WriteString("        type: apiKey\n")
	f.WriteString("        in: cookie\n")
	f.WriteString("        name: SESSIONID\n")
	f.WriteString("  schemas:\n")
	f.WriteString("    Error:\n")
	f.WriteString("      type: object\n")
	f.WriteString("      properties:\n")
	f.WriteString("        code:\n")
	f.WriteString("          type: string\n")
	f.WriteString("        message:\n")
	f.WriteString("          type: string\n")
	f.WriteString("      required:\n")
	f.WriteString("        - code\n")
	f.WriteString("        - message\n")

	f.WriteString(swagger.SwaggerStructsWithSuffix(fset, pkgs, "Request", 4))
	f.WriteString(swagger.SwaggerStructsWithSuffix(fset, pkgs, "Response", 4))

	f.WriteString("  responses:\n")
	f.WriteString("    pongResponse:\n")
	f.WriteString("      description: Server is up and running\n")
	f.WriteString("      content:\n")
	f.WriteString("        text/plain:\n")
	f.WriteString("          schema:\n")
	f.WriteString("            type: string\n")
	f.WriteString("            example: pong\n")
	f.WriteString("    ZipFile:\n")
	f.WriteString("      description: A file as a download.\n")
	f.WriteString("      content:\n")
	f.WriteString("        application/zip:\n")
	f.WriteString("          schema:\n")
	f.WriteString("            type: string\n")
	f.WriteString("            format: binary\n")
	f.WriteString("    ImageFile:\n")
	f.WriteString("      description: A file as a download.\n")
	f.WriteString("      content:\n")
	f.WriteString("        image/jpeg:\n")
	f.WriteString("          schema:\n")
	f.WriteString("            type: string\n")
	f.WriteString("            format: binary\n")
	f.WriteString("    OK:\n")
	f.WriteString("      description: Post successfully delivered.\n")
	f.WriteString("    NoContent:\n")
	f.WriteString("      description: Update was successful.\n")
	f.WriteString("    BadRequest:\n")
	f.WriteString("      description: The request is in a wrong format or contains missing fields.\n")
	f.WriteString("      content:\n")
	f.WriteString("        application/json:\n")
	f.WriteString("          schema:\n")
	f.WriteString("            $ref: \"#/components/schemas/Error\"\n")
	f.WriteString("    Unauthenticated:\n")
	f.WriteString("      description: User is not logged in.\n")
	f.WriteString("      content:\n")
	f.WriteString("        application/json:\n")
	f.WriteString("          schema:\n")
	f.WriteString("            $ref: \"#/components/schemas/Error\"\n")
	f.WriteString("    Unauthorized:\n")
	f.WriteString("      description: User is logged in but has not the permission to perform the request.\n")
	f.WriteString("      content:\n")
	f.WriteString("        application/json:\n")
	f.WriteString("          schema:\n")
	f.WriteString("            $ref: \"#/components/schemas/Error\"\n")

	// create all responses
	f.WriteString(swagger.SwaggerResponsesWithSuffix(fset, pkgs, "Response", 4))

	duplicateResponseLists := make(map[string]int)

	// create all list responses
	pre := strings.Repeat(" ", 4)
	for url := range endpoints {
		for _, action := range endpoints[url] {
			for _, r := range action.Details.Responses {
				text := strings.TrimSpace(r.Text)
				if strings.HasSuffix(text, "List") {

					_, exists := duplicateResponseLists[strings.TrimSpace(text)]

					if !exists {
						f.WriteString(fmt.Sprintf("%s%s:\n", pre, text))
						f.WriteString(fmt.Sprintf("%s  description: done\n", pre))
						f.WriteString(fmt.Sprintf("%s  content:\n", pre))
						f.WriteString(fmt.Sprintf("%s    application/json:\n", pre))
						f.WriteString(fmt.Sprintf("%s      schema:\n", pre))
						f.WriteString(fmt.Sprintf("%s        type: array\n", pre))
						f.WriteString(fmt.Sprintf("%s        items:\n", pre))
						f.WriteString(fmt.Sprintf("%s          $ref: \"#/components/schemas/%s\"\n", pre, text[:len(text)-4]))

						duplicateResponseLists[strings.TrimSpace(text)] = 0
					}
				}

			}
		}
	}

	f.WriteString("paths:\n")
	for url := range endpoints {

		f.WriteString(fmt.Sprintf("  %s:\n", url))
		for _, action := range endpoints[url] {

			if action.Details.Method == "post" || action.Details.Method == "put" || action.Details.Method == "patch" {
				if action.Details.Request == "" {
					panic(fmt.Sprintf("endpoint '%s' is '%s' but has no request body in %v",
						url, action.Details.Method, action.Position))
				}
			}

			if action.Details.Method == "get" {
				// test wether we have a 200 response
				found := false
				for _, r := range action.Details.Responses {
					if r.Code == 200 {
						found = true
						break
					}
				}
				if !found {
					panic(fmt.Sprintf("endpoint '%s' is '%s' but has no 200 response in %v",
						url, action.Details.Method, action.Position))
				}
			}

			f.WriteString(fmt.Sprintf("    # implementation in  %v\n", action.Position))
			f.WriteString(fmt.Sprintf("    %s:\n", action.Details.Method))
			f.WriteString(fmt.Sprintf("      summary: %s\n", action.Details.Summary))
			if action.Details.Description != "" {
				f.WriteString(fmt.Sprintf("      description: >\n      %s\n", action.Details.Description))
			}
			if len(action.Details.Tags) > 0 {
				f.WriteString("      tags: \n\n")
				for _, tag := range action.Details.Tags {

					f.WriteString(fmt.Sprintf("        - %s\n", tag))
				}
			}
			if len(action.Details.URLParams)+len(action.Details.QueryParams) > 0 {
				f.WriteString("      parameters:\n")
				for _, el := range action.Details.URLParams {
					f.WriteString("        - in: path\n")
					f.WriteString(fmt.Sprintf("          name: %s\n", el.Name))
					f.WriteString("          schema: \n")
					f.WriteString(fmt.Sprintf("             type: %s\n", el.Type))
					f.WriteString("          required: true \n")
				}

				for _, el := range action.Details.QueryParams {
					f.WriteString("        - in: query\n")
					f.WriteString(fmt.Sprintf("          name: %s\n", el.Name))
					f.WriteString("          schema: \n")
					f.WriteString(fmt.Sprintf("             type: %s\n", el.Type))
					f.WriteString("          required: false \n")
				}
			}

			if action.Details.Request != "" {
				f.WriteString("      requestBody:\n")
				f.WriteString("        required: true\n")

				switch action.Details.Request {
				case "zipfile":
					f.WriteString("        content:\n")
					f.WriteString("          multipart/form-data:\n")
					f.WriteString("            schema:\n")
					f.WriteString("              type: object\n")
					f.WriteString("              properties:\n")
					f.WriteString("                file_data:\n")
					f.WriteString("                  type: string\n")
					f.WriteString("                  format: binary\n")
					f.WriteString("            encoding:\n")
					f.WriteString("              file_data:\n")
					f.WriteString("                contentType: application/zip\n")
				case "imagefile":
					f.WriteString("        content:\n")
					f.WriteString("          multipart/form-data:\n")
					f.WriteString("            schema:\n")
					f.WriteString("              type: object\n")
					f.WriteString("              properties:\n")
					f.WriteString("                file_data:\n")
					f.WriteString("                  type: string\n")
					f.WriteString("                  format: binary\n")
					f.WriteString("            encoding:\n")
					f.WriteString("              file_data:\n")
					f.WriteString("                contentType: image/jpeg\n")
				case "empty":

				default:
					f.WriteString("        content:\n")
					f.WriteString("          application/json:\n")
					f.WriteString("            schema:\n")
					f.WriteString(fmt.Sprintf("              $ref: \"#/components/schemas/%s\"\n", action.Details.Request))
				}

			}
			f.WriteString("      responses:\n")
			for _, r := range action.Details.Responses {

				f.WriteString(fmt.Sprintf("        \"%v\":\n", r.Code))
				f.WriteString(fmt.Sprintf("          $ref: \"#/components/responses/%s\"\n", r.Text))

			}
		}

	}

	f.Sync()

}
