# InfoMark-Backend

[![Build Status](https://ci.patwie.com/api/badges/cgtuebingen/infomark-backend/status.svg)](http://ci.patwie.com/cgtuebingen/infomark-backend)

InfoMark is a CI inspired online course management system. The goal is to achieve auto testing of exercises/homework using unit tests to ease the task of TAs.
This repo hosts the backend of the application. It is written in [Go](https://golang.org/). The API is defined in this [repository](https://github.com/cgtuebingen/infomark-swagger)
using [Swagger](https://swagger.io/).

The frontend is implemented in [Elm]((https://elm-lang.org/)) and is available [here](https://github.com/cgtuebingen/infomark-frontend).


# Quick Setup for Demo Purposes

No manual settings, just use the defaults (just for development).

```bash
git clone https://github.com/cgtuebingen/infomark-backend.git
go build infomark.go
cp .informark.yml.example ~/.informark.yml

sudo docker-compose up -d

cd database
python3 mock.py
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f schema.sql

# you might want o skip this step (in production)
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f mock.sql
cd ..

./infomark-backend server

# debug database
pgweb --host=127.0.0.1 --port 5433 --user=user --pass=pass --db=db
```


# More Details

The backend is written in GO and acts as a REST server processing request and sending responses (`./infomark serve`). See below to generate the Swagger definition file automatically. To test submissions multiple separate processes can be spawn like `./infomark work`. They will use RabbitMQ to communicate with the server. More precisely the server will enqueue the files whenever a student uploads a solution. In round-robin fashion the workers will download the files and test them in a docker environment sending the result back via a HTTP request. (There might be race-condition when fetching from worker and uploading a new file from the student occurs on the same time).

The is a single configuration file "~/.infomark.yml" in the home directory of the user. This will contain information for both "server" and "worker".

# Production Setup

In this setup, we need to edit all configuration files. I suggest to use the `docker-compose.yml` to run dependencies but run infomark outside of docker as it is a single binary file. Whenever your config ask for a password or secret use

```bash
openssl rand -hex 32
```

to generate a random one.

We will edit this time the configfile

```bash
git clone https://github.com/cgtuebingen/infomark-backend.git
go build infomark.go
cp .informark.yml.example ~/.informark.yml
edit ~/.informark.yml
```

If you want to install dependencies on your own, feel free as informark only expects the connection string for dependencies. See [docs/INSTALL.md](./docs/INSTALL.md) for more details. However, it should be sufficient to use the `docker-compose.yml` file. If you are running multiple postgres services, make sure you use different ports.


# Development Setup

During development I also use `docker-compose.yml` for simplicity. For running the unit tests and manually debug the REST endpoints, fake entries are created for the database (faking users with names, submissions and task). IMPORTANT: If you are working on the [mock-generator](./database/mock.py) please add new lines to the bottom as the random seed is fixed and some unit tests depend on the actual random seed.

To run the tests or having actual data to display, we use some mocking generate in python:

```bash
# do it once to create schema and mock
cd database
python3 mock.py
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f schema.sql
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f mock.sql
cd ..
```

If during development you want to reset everything just do

```bash
sudo docker-compose down -v
sudo docker-compose rm
sudo docker-compose up --force-recreate
```

To debug the database I suggest to use [pgweb](https://github.com/sosedoff/pgweb)

```bash
pgweb --host=127.0.0.1 --port 5433 --user=user --pass=pass --db=db
```

## Testing

We ship unit tests and a database mock which is generated in Python. Running the tests is mandatory to ensure stability and correctness and we suggest to at least these tests once locally.

```bash
cd api/app
go test -cover
```

## Generating the Docs

This implementation features an automatic Swagger-v3 definition generation of all available endpoints.
Further, generating this `api.yaml` file also checks if all routes are documented and have the correct method (get, post, patch, put) and if responses or request bodies are missing. Further, it uses the request and response go-structs to generate request and response body information for swagger automatically.

Just run

```bash
go generate
```


