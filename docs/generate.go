package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cgtuebingen/infomark-backend/api/app"
	"github.com/cgtuebingen/infomark-backend/docs/swagger"
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

type Routes struct {
	Method string
	Path   string
}

func GetAllRoutes() []*Routes {
	db, _ := sqlx.Open("sqlite3", ":memory:")
	r, _ := app.New(db, false)

	routes := []*Routes{}

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		route = strings.Replace(route, "/*/", "/", -1)
		route = strings.Replace(route, "/*/", "/", -1)
		route = strings.Replace(route, "/*/", "/", -1)
		route = strings.Replace(route, "/*", "/", -1)
		// f.WriteString(fmt.Sprintf("%s %s\n", method, route))

		if len(route) > 1 && strings.HasPrefix(route, "/api/v1") {

			if strings.HasSuffix(route, "/") {
				route = route[:len(route)-1]
			}

			routes = append(routes, &Routes{Method: strings.ToLower(method), Path: route[7:]})

		}
		return nil
	}

	if err := chi.Walk(r, walkFunc); err != nil {
		panic(fmt.Sprintf("Logging err: %s\n", err.Error()))
	}

	return routes
}

var cfgFile = ""

func SetConfigFile() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		var err error
		// Find home directory.
		home := os.Getenv("INFOMARK_CONFIG_DIR")

		if home == "" {
			home, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
		}

		// Search config in home directory with name ".go-base" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".infomark")
	}

}

// Main docs
func main() {

	SetConfigFile()
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	fset := token.NewFileSet() // positions are relative to fset

	pkgs, err := parser.ParseDir(fset, "./api/app/", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	endpoints := swagger.GetEndpoints(pkgs, fset)

	// verify we have all api-routes in the swagger description
	routes := GetAllRoutes()
	for _, route := range routes {
		found := false
		for url := range endpoints {
			if route.Path == url {

				for _, action := range endpoints[url] {
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

					panic(fmt.Sprintf("found '%s' but not the correct method want '%s' \n candiates are %v", url, route.Method, candidates))
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

	f.WriteString(fmt.Sprintf("# this file has been automatically created. Please do not edit it here\n"))
	f.WriteString(fmt.Sprintf("openapi: 3.0.0\n"))
	f.WriteString(fmt.Sprintf("info:\n"))
	f.WriteString(fmt.Sprintf("  title: \"InfoMark\"\n"))
	f.WriteString(fmt.Sprintf("  version: \"0.0.1\"\n"))
	f.WriteString(fmt.Sprintf("  description: >\n"))
	f.WriteString(fmt.Sprintf("    A CI based course framework. All enums should be send as strings and returned as strings.\n"))
	f.WriteString(fmt.Sprintf("    Everything\n"))
	f.WriteString(fmt.Sprintf("  contact:\n"))
	f.WriteString(fmt.Sprintf("    - name: Mark Boss\n"))
	f.WriteString(fmt.Sprintf("      email: mark.boss@uni-tuebingen.de\n"))
	f.WriteString(fmt.Sprintf("      url: https://uni-tuebingen.de\n"))
	f.WriteString(fmt.Sprintf("    - name: Patrick Wieschollek\n"))
	f.WriteString(fmt.Sprintf("      email: Patrick.Wieschollek@uni-tuebingen.de\n"))
	f.WriteString(fmt.Sprintf("      url: https://uni-tuebingen.de\n"))
	f.WriteString(fmt.Sprintf("servers:\n"))
	f.WriteString(fmt.Sprintf("  - url: http://localhost:3000/api/v1\n"))
	f.WriteString(fmt.Sprintf("security:\n"))
	f.WriteString(fmt.Sprintf("  - bearerAuth: []\n"))
	f.WriteString(fmt.Sprintf("  - cookieAuth: []\n"))
	f.WriteString(fmt.Sprintf("tags:\n"))
	f.WriteString(fmt.Sprintf("  - name: common\n"))
	f.WriteString(fmt.Sprintf("    description: common request\n"))
	f.WriteString(fmt.Sprintf("  - name: auth\n"))
	f.WriteString(fmt.Sprintf("    description: authenticated related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: account\n"))
	f.WriteString(fmt.Sprintf("    description: account related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: email\n"))
	f.WriteString(fmt.Sprintf("    description: Email related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: users\n"))
	f.WriteString(fmt.Sprintf("    description: User related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: courses\n"))
	f.WriteString(fmt.Sprintf("    description: Course related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: sheets\n"))
	f.WriteString(fmt.Sprintf("    description: Exercise sheets related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: tasks\n"))
	f.WriteString(fmt.Sprintf("    description: Exercise tasks related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: submissions\n"))
	f.WriteString(fmt.Sprintf("    description: Submissions related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: grades\n"))
	f.WriteString(fmt.Sprintf("    description: Gradings related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: groups\n"))
	f.WriteString(fmt.Sprintf("    description: Exercise groups related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: enrollments\n"))
	f.WriteString(fmt.Sprintf("    description: Enrollments related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: materials\n"))
	f.WriteString(fmt.Sprintf("    description: Exercise material related requests\n"))
	f.WriteString(fmt.Sprintf("  - name: internal\n"))
	f.WriteString(fmt.Sprintf("    description: Endpoints for internal usage only\n"))

	f.WriteString(fmt.Sprintf("components:\n"))
	f.WriteString(fmt.Sprintf("  securitySchemes:\n"))
	f.WriteString(fmt.Sprintf("      bearerAuth:\n"))
	f.WriteString(fmt.Sprintf("        type: http\n"))
	f.WriteString(fmt.Sprintf("        scheme: bearer\n"))
	f.WriteString(fmt.Sprintf("        bearerFormat: JWT\n"))
	f.WriteString(fmt.Sprintf("      cookieAuth:\n"))
	f.WriteString(fmt.Sprintf("        type: apiKey\n"))
	f.WriteString(fmt.Sprintf("        in: cookie\n"))
	f.WriteString(fmt.Sprintf("        name: SESSIONID\n"))
	f.WriteString(fmt.Sprintf("  schemas:\n"))
	f.WriteString(fmt.Sprintf("    Error:\n"))
	f.WriteString(fmt.Sprintf("      type: object\n"))
	f.WriteString(fmt.Sprintf("      properties:\n"))
	f.WriteString(fmt.Sprintf("        code:\n"))
	f.WriteString(fmt.Sprintf("          type: string\n"))
	f.WriteString(fmt.Sprintf("        message:\n"))
	f.WriteString(fmt.Sprintf("          type: string\n"))
	f.WriteString(fmt.Sprintf("      required:\n"))
	f.WriteString(fmt.Sprintf("        - code\n"))
	f.WriteString(fmt.Sprintf("        - message\n"))

	f.WriteString(swagger.SwaggerStructsWithSuffix(fset, pkgs, "Request", 4))
	f.WriteString(swagger.SwaggerStructsWithSuffix(fset, pkgs, "Response", 4))

	f.WriteString(fmt.Sprintf("  responses:\n"))
	f.WriteString(fmt.Sprintf("    pongResponse:\n"))
	f.WriteString(fmt.Sprintf("      description: Server is up and running\n"))
	f.WriteString(fmt.Sprintf("      content:\n"))
	f.WriteString(fmt.Sprintf("        text/plain:\n"))
	f.WriteString(fmt.Sprintf("          schema:\n"))
	f.WriteString(fmt.Sprintf("            type: string\n"))
	f.WriteString(fmt.Sprintf("            example: pong\n"))
	f.WriteString(fmt.Sprintf("    ZipFile:\n"))
	f.WriteString(fmt.Sprintf("      description: A file as a download.\n"))
	f.WriteString(fmt.Sprintf("      content:\n"))
	f.WriteString(fmt.Sprintf("        application/zip:\n"))
	f.WriteString(fmt.Sprintf("          schema:\n"))
	f.WriteString(fmt.Sprintf("            type: string\n"))
	f.WriteString(fmt.Sprintf("            format: binary\n"))
	f.WriteString(fmt.Sprintf("    ImageFile:\n"))
	f.WriteString(fmt.Sprintf("      description: A file as a download.\n"))
	f.WriteString(fmt.Sprintf("      content:\n"))
	f.WriteString(fmt.Sprintf("        image/jpeg:\n"))
	f.WriteString(fmt.Sprintf("          schema:\n"))
	f.WriteString(fmt.Sprintf("            type: string\n"))
	f.WriteString(fmt.Sprintf("            format: binary\n"))
	f.WriteString(fmt.Sprintf("    OK:\n"))
	f.WriteString(fmt.Sprintf("      description: Post successfully delivered.\n"))
	f.WriteString(fmt.Sprintf("    NoContent:\n"))
	f.WriteString(fmt.Sprintf("      description: Update was successful.\n"))
	f.WriteString(fmt.Sprintf("    BadRequest:\n"))
	f.WriteString(fmt.Sprintf("      description: The request is in a wrong format or contains missing fields.\n"))
	f.WriteString(fmt.Sprintf("      content:\n"))
	f.WriteString(fmt.Sprintf("        application/json:\n"))
	f.WriteString(fmt.Sprintf("          schema:\n"))
	f.WriteString(fmt.Sprintf("            $ref: \"#/components/schemas/Error\"\n"))
	f.WriteString(fmt.Sprintf("    Unauthenticated:\n"))
	f.WriteString(fmt.Sprintf("      description: User is not logged in.\n"))
	f.WriteString(fmt.Sprintf("      content:\n"))
	f.WriteString(fmt.Sprintf("        application/json:\n"))
	f.WriteString(fmt.Sprintf("          schema:\n"))
	f.WriteString(fmt.Sprintf("            $ref: \"#/components/schemas/Error\"\n"))
	f.WriteString(fmt.Sprintf("    Unauthorized:\n"))
	f.WriteString(fmt.Sprintf("      description: User is logged in but has not the permission to perform the request.\n"))
	f.WriteString(fmt.Sprintf("      content:\n"))
	f.WriteString(fmt.Sprintf("        application/json:\n"))
	f.WriteString(fmt.Sprintf("          schema:\n"))
	f.WriteString(fmt.Sprintf("            $ref: \"#/components/schemas/Error\"\n"))

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

	f.WriteString(fmt.Sprintf("paths:\n"))
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
				f.WriteString(fmt.Sprintf("      tags: \n\n"))
				for _, tag := range action.Details.Tags {

					f.WriteString(fmt.Sprintf("        - %s\n", tag))
				}
			}
			if len(action.Details.URLParams)+len(action.Details.QueryParams) > 0 {
				f.WriteString(fmt.Sprintf("      parameters:\n"))
				for _, el := range action.Details.URLParams {
					f.WriteString(fmt.Sprintf("        - in: path\n"))
					f.WriteString(fmt.Sprintf("          name: %s\n", el.Name))
					f.WriteString(fmt.Sprintf("          schema: \n"))
					f.WriteString(fmt.Sprintf("             type: %s\n", el.Type))
					f.WriteString(fmt.Sprintf("          required: true \n"))
				}

				for _, el := range action.Details.QueryParams {
					f.WriteString(fmt.Sprintf("        - in: query\n"))
					f.WriteString(fmt.Sprintf("          name: %s\n", el.Name))
					f.WriteString(fmt.Sprintf("          schema: \n"))
					f.WriteString(fmt.Sprintf("             type: %s\n", el.Type))
					f.WriteString(fmt.Sprintf("          required: false \n"))
				}
			}

			if action.Details.Request != "" {
				f.WriteString(fmt.Sprintf("      requestBody:\n"))
				f.WriteString(fmt.Sprintf("        required: true\n"))

				switch action.Details.Request {
				case "zipfile":
					f.WriteString(fmt.Sprintf("        content:\n"))
					f.WriteString(fmt.Sprintf("          multipart/form-data:\n"))
					f.WriteString(fmt.Sprintf("            schema:\n"))
					f.WriteString(fmt.Sprintf("              type: object\n"))
					f.WriteString(fmt.Sprintf("              properties:\n"))
					f.WriteString(fmt.Sprintf("                file_data:\n"))
					f.WriteString(fmt.Sprintf("                  type: string\n"))
					f.WriteString(fmt.Sprintf("                  format: binary\n"))
					f.WriteString(fmt.Sprintf("            encoding:\n"))
					f.WriteString(fmt.Sprintf("              file_data:\n"))
					f.WriteString(fmt.Sprintf("                contentType: application/zip\n"))
				case "imagefile":
					f.WriteString(fmt.Sprintf("        content:\n"))
					f.WriteString(fmt.Sprintf("          multipart/form-data:\n"))
					f.WriteString(fmt.Sprintf("            schema:\n"))
					f.WriteString(fmt.Sprintf("              type: object\n"))
					f.WriteString(fmt.Sprintf("              properties:\n"))
					f.WriteString(fmt.Sprintf("                file_data:\n"))
					f.WriteString(fmt.Sprintf("                  type: string\n"))
					f.WriteString(fmt.Sprintf("                  format: binary\n"))
					f.WriteString(fmt.Sprintf("            encoding:\n"))
					f.WriteString(fmt.Sprintf("              file_data:\n"))
					f.WriteString(fmt.Sprintf("                contentType: image/jpeg\n"))
				case "empty":

				default:
					f.WriteString(fmt.Sprintf("        content:\n"))
					f.WriteString(fmt.Sprintf("          application/json:\n"))
					f.WriteString(fmt.Sprintf("            schema:\n"))
					f.WriteString(fmt.Sprintf("              $ref: \"#/components/schemas/%s\"\n", action.Details.Request))
				}

			}
			f.WriteString(fmt.Sprintf("      responses:\n"))
			for _, r := range action.Details.Responses {

				f.WriteString(fmt.Sprintf("        \"%v\":\n", r.Code))
				f.WriteString(fmt.Sprintf("          $ref: \"#/components/responses/%s\"\n", r.Text))

			}
		}

	}

	f.Sync()

}
