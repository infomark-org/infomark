# InfoMark-Backend

[![Build Status](https://ci.patwie.com/api/badges/cgtuebingen/infomark-backend/status.svg)](http://ci.patwie.com/cgtuebingen/infomark-backend)

InfoMark is a CI inspired online course management system. The goal is to achieve auto testing of exercises/homework using unit tests to ease the task of TAs.
This repo hosts the backend of the application. It is written in [Go](https://golang.org/). The API is defined in this [repository](https://github.com/cgtuebingen/infomark-swagger)
using [Swagger](https://swagger.io/).

The frontend is implemented in [Elm]((https://elm-lang.org/)) and is available [here](https://github.com/cgtuebingen/infomark-frontend).


# Quick Setup

No manual settings, just use the defaults (just for development).

```bash
git clone https://github.com/cgtuebingen/infomark-backend.git
go build infomark-backend.go
cp .informark-backend.yml.example ~/.informark-backend.yml

sudo docker-compose up -d

cd database
python3 mock.py
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f schema.sql
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f mock.sql
cd ..

./infomark-backend server
```


# Production Setup

We will edit this time the configfile

```bash
git clonehttps://github.com/cgtuebingen/infomark-backend.git
go build infomark-backend.go
cp .informark-backend.yml.example ~/.informark-backend.yml
edit ~/.informark-backend.yml
```

Note, you should use

```bash
openssl rand -hex 32
```

to generate random passwords or secrets.

# Development Setup
## Mocking

To run the tests or having actual data to display, we use some mocking generate in python:

```bash
# do it once to create schema and mock
cd database
python3 mock.py
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f schema.sql
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f mock.sql
cd ..
```

Reminder, to reset the docker-compose files (chaning password) just do

```bash
sudo docker-compose down -v
sudo docker-compose rm
sudo docker-compose up --force-recreate
```

## Testing

We ship unit tests and a database mock which is generated in Python. Read the docs [docs](./docs/) for more details of how to use a database mock. Running the tests is mandatory to ensure stability and correctness and we suggest to at least these tests once locally.



## Generating the Docs

The command

```bash
go generate
```

will generate a valid `api.yaml` for Swagger 3.0.
Hereby, it verifies all implemented routes are documented and have the correct method (get, post, patch, put).
Further, it uses the request and response go-structs to generate request and response body information in swagger.