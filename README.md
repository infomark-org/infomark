# InfoMark

[![Build Status](https://ci.patwie.com/api/badges/cgtuebingen/infomark-backend/status.svg)](http://ci.patwie.com/cgtuebingen/infomark-backend)

InfoMark is an is a scalable, modern and open-source [rewrite of our](https://github.com/cgtuebingen/InfoMark-deprecated)
online course management system with auto testing of students submissions using unit tests to ease the task of TAs.

Features:
- flexible client/server implementation featuring unit-tests
- distribute exercise sheets with due-dates, and course slides/material with publish-dates
- students can upload their solutions
- assignments of students to exercise groups according their bids is optimized via MILP solver
- automatic asynchronous testing of students homework solutions by scalable background workers using docker as a sandbox and providing feedback for students
- easy to install using docker-compose for dependencies and single binary for the server
- CLI for administrative work without touching the database

The frontend is Single-Page-Applications written in [Elm]((https://elm-lang.org/)) based on REST-backend server written in [Go](https://golang.org/). Communication to internal background workers uses [RabbitMQ](https://www.rabbitmq.com/) and data is stored in a [PostgreSQL](https://www.postgresql.org/) database. The [Swagger](https://swagger.io/) definitions for the REST endpoints can be automatically generated from the REST server. It is possible to export/import the [Symphony](https://projects.coin-or.org/SYMPHONY) problem/solution format.

This gives all advantages:
- Cross-platform: backend runs on OSX, Linux and Windows
- Easy to Install: a single binary contains the entire backend
- Minimal Dependencies: just RabbitMQ and PostgreSQL is required
- Highly-Scalable: background workers can be deployed on different machines and the unit-tests will be distributed amongst them

## Quick Setup for Demo Purposes

No manual settings, plug and play:

```bash
# get code
git clone https://github.com/cgtuebingen/infomark-backend.git
go build infomark.go
# copy pre-defined config
cp .informark.yml.example ~/.informark.yml
# start dependencies (postgres, rabbitmq)
sudo docker-compose up -d

# initialize database
cd database
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f schema.sql
# generate some dummy data
python3 mock.py
PGPASSWORD=pass psql -h 127.0.0.1 -U user -p 5433 -d db -f mock.sql
cd ..

# start a single background worker (feel free to start more instanced)
# as it uses docker, sudo permissions are required to talk to the docker context
sudo ./infomark worker &

# start server
./infomark serve

# generate swagger
go generate
```

Point your browser to `http://localhost:3000` and use the credentials:

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

and use the Swagger-UI.


## Architecture + Philosophy

The system uses a separate UI (written in ELM) and REST backend (written in GO). Every HTTP-request from a frontend requires either a cookie-based session or JWT header-token.

All involved features are optimized to ensure fairness. This starts by assigning students to an exercise group according their bids optimizing the global happiness using [Symphony](https://projects.coin-or.org/SYMPHONY) (see below).

Course administrators create exercise sheets with deadlines and unit-tests. Students can upload their homework solutions in a given period of time. The server will block submissions after the deadline for that task.

These uploads are automatically distributed across background workers (using RabbitMQ) which employ docker to run two unit tests against the uploads in a sandbox: a public unit test and private unit test. The public unit test will provide the student immediate feedback. The private unit tests are just visible to TAs (tutors, administrators) to avoid 'overfitting' of the public unit tests. This is language independent (we provide an example for JAVA).

A internal cronjob will zip all submissions after the deadline, so that TAs can download these submissions in one go. TAs can see the result of the private unit tests and grade the submitted solutions of the students. The results of the unit tests is optional and should just guide the TAs.

## Development + Unit-Tests

Some tools have been proven useful during development. We suggest [pgweb](https://github.com/sosedoff/pgweb) to visualize the database queries

```bash
# debug database
pgweb --host=127.0.0.1 --port 5433 --user=user --pass=pass --db=db
```

To run the unit-tests you will need to mock the database (we need some data to test again) in python:

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

If you are working on the [mock-generator](./database/mock.py) please add new lines to the bottom as the random seed is fixed and some unit tests depend on the actual random seed.

# Production Setup

In this setup, we need to edit the configuration files. We still suggest to use the `docker-compose.yml` to run dependencies (RabbitMQ, Postgres) but run the infomark server outside of docker as it is a single binary file. Whenever your config ask for a password or secret use

```bash
openssl rand -hex 32
```

to generate a random one.

We will edit this time the config-file

```bash
edit ~/.informark.yml
```

We suggest to put the REST-server behind an [NGINX](https://www.nginx.com/) instance using our [config file](./external/nginx/infomark.conf).

## Infomark CLI

The infomark binary has several commands in-built to activate users, trigger a unit test run again, export/import bids/assignments of students-to-groups.

### Group Assignment

Using the informark cli console, we can export (anonymously) the bids of each student for eachg group by

```bash
export COURSEID = 1
export FILENAME = test2
export MIN_PER_GROUP = 28
export MAX_PER_GROUP = 35

./infomark console assignments dump_bids ${COURSEID} ${FILENAME} ${MIN_PER_GROUP} ${MAX_PER_GROUP}

# solve the MILP optimally
sudo docker run -v "$PWD":/data -it patwie/symphony  /var/symphony/bin/symphony -F test2.mod -D test2.dat -f test2.par > solution.txt


./infomark console assignments import-solution ${COURSEID} solution.txt
```