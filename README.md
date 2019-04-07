# InfoMark

[![Build Status](https://ci.patwie.com/api/badges/cgtuebingen/infomark-backend/status.svg)](http://ci.patwie.com/cgtuebingen/infomark-backend)

InfoMark is an is a scalable, modern and open-source [rewrite of our](https://github.com/cgtuebingen/InfoMark-deprecated)
online course management system with auto testing of students submissions using unit tests to ease the task of TAs.

For more details see [https://github.com/cgtuebingen/infomark](https://github.com/cgtuebingen/infomark).

## Quick Setup for Development

No manual settings, plug and play:

```bash
# get code
git clone https://github.com/cgtuebingen/infomark-backend.git
go build infomark.go

# copy pre-defined config (just for demo purposes)
cp .informark.yml.example .informark.yml
# the backend will search for .infomark.yml in the directory $INFOMARK_CONFIG_DIR
# or use the file in the current directory

# start dependencies (postgres, rabbitmq, redis)
sudo docker-compose up -d

# initialize database
cd database
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f schema.sql
# generate some dummy data
python3 mock.py
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f mock.sql
cd ..

# start a single background worker (feel free to start more instances)
# as it uses docker, sudo permissions are required to talk to the docker context
sudo ./infomark worker &
sudo ./infomark worker &
sudo ./infomark worker &

# start server
./infomark serve

# generate swagger
go generate
```

The backend will be served at `http://localhost:3000` and some credentials are:

```
user: test@uni-tuebingen.de
pass: test
```

## Generating the Docs (Swagger)

This implementation features an automatic Swagger-v3 definition generation of all available endpoints from the REST-endpoints which includes code-checks to ensure that:
- all routes are documented and have the correct method (GET, POST, PATCH, PUT)
- responses or request bodies have documentations and examples

Just run

```bash
go generate
```

and use the Swagger-UI to serve the file `api.yaml`.



## Development + Unit-Tests

Some tools have been proven useful during development. We suggest [pgweb](https://github.com/sosedoff/pgweb) to debug the database queries

```bash
pgweb --host=127.0.0.1 --port 5433 --user=user --pass=pass --db=db
```

To run the unit-tests you will need to mock the database content (we need some data to test again) in python:

```bash
# do it once to create schema and mock
cd database
python3 mock.py
# reset database
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f schema.sql
# generate some data
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f mock.sql
cd ..

cd api/app
# this is run by our CI tests
go test -cover
```

If during development (and using docker-compose) you want to reset everything just do

```bash
sudo docker-compose down -v
sudo docker-compose rm
sudo docker-compose up --force-recreate
```

If you are working on the [mock-generator](./database/mock.py) please add new lines to the end of the file as the random seed is fixed and some unit tests depend on the actual random seed.
