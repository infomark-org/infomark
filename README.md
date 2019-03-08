# InfoMark-Backend

[![Build Status](https://ci.patwie.com/api/badges/cgtuebingen/infomark-backend/status.svg)](http://ci.patwie.com/cgtuebingen/infomark-backend)

InfoMark is a CI inspired online course management system. The goal is to achieve auto testing of exercises/homework using unit tests to ease the task of TAs.
This repo hosts the backend of the application. It is written in [Go](https://golang.org/). The API is defined in this [repository](https://github.com/cgtuebingen/infomark-swagger)
using [Swagger](https://swagger.io/).

The frontend is implemented in [Elm]((https://elm-lang.org/)) and is available [here](https://github.com/cgtuebingen/infomark-frontend).

# Building

Please have a look at the [`.drone.yml`](./.drone.yml) config for more details.

```bash
git clone <this repo>
go build infomark-backend.go
cp .informark-backend.yml.example ~/.informark-backend.yml
edit ~/.informark-backend.yml
```

# Testing

We ship unit tests and a database mock which is generated in Python. Read the docs [docs](./docs/) for more details of how to use a database mock. Running the tests is mandatory to ensure stability and correctness and we suggest to at least these tests once locally.


# Running

```bash
./infomark-backend server
```

# Generating the Docs

The command

```bash
go generate
```

will generate a valid `api.yaml` for Swagger 3.0.
Hereby, it verifies all implemented routes are documented and have the correct method (get, post, patch, put).
Further, it uses the request and response go-structs to generate request and response body information in swagger.