# Overview

InfoMark is a free, scalable, modern, and open-source online course
management system supporting automatic testing of programming assignments,
scaling to thousands of students across several courses.

Uploaded solutions to programming assignments are tested automatically. For
more information on how to write such tests, see the [Tutor's Guide](tutor.md).
To learn how to use the system, refer to the
[Administrator's Guide](administrator.md). For development, refer to the
[Developer's Guide](developer.md).

Teaching assistants (tutors) grade these homework solutions online. The
platform supports multiple courses, each with several exercise groups, slides,
and course material. The backend server speaks RESTful JSON, so you can write
your own scripts against it in, for example, Python.

# Quick start

These commands are the *same* whether you deploy on your local machine or in
production on a server. Please download the latest release from the
[release page](https://github.com/infomark-org/infomark/releases/). Each
release ships a single binary containing all required files, including the
compiled web interface. The only dependency is Docker and Docker Compose.

The following steps spin up a fully production-ready system from scratch.
InfoMark is implemented as a modern CLI with POSIX-compliant flags.

Without any lengthy explanation, you can start an instance of InfoMark from
scratch with:

```bash
# Create configuration
./infomark console configuration create > infomark-config.yml
# Create docker-compose for database, rabbitMQ, redis
./infomark console configuration create-compose infomark-config.yml > docker-compose.yml
# Start dependencies
sudo docker-compose up -d
# Start InfoMark server
export INFOMARK_CONFIG_FILE=`realpath infomark-config.yml`
./infomark serve
# point your browser to http://localhost:2020
```

## Requirements

InfoMark has the following minimal requirements:

* one CPU core for the server `infomark serve` (1 GB RAM)
* one CPU core for each background worker `infomark work` (the memory
  requirement depends on the size of the Docker image used for your
  programming assignments)

We assume an Ubuntu system, but InfoMark will happily run on any other system
that provides Docker (>= v1.13) and Docker Compose. There are no other
dependencies to juggle.

## Setup

First, create a configuration with InfoMark and write it to
`infomark-config.yml`.

```bash
./infomark console configuration create > infomark-config.yml
```

This config file is populated with values for a minimal working setup
([example](https://github.com/infomark-org/infomark/blob/master/configuration/example.yml)).
Strong passwords are generated (each time you call this command). The
configuration file might seem a bit complex at first glance, but it should
work out of the box.

We use Docker Compose to handle the dependencies in a sandbox:

```bash
./infomark console configuration create-compose infomark-config.yml > docker-compose.yml
```

This creates a ready-to-use docker-compose file. All of the following commands
require the configuration file `infomark-config.yml`, which they expect to
find via the environment variable `INFOMARK_CONFIG_FILE`. You can point to it
with:

```bash
export INFOMARK_CONFIG_FILE=/absolute/path/to/infomark-config.yml
```

If you forget to set this environment variable, the following commands will
remind you to do so.

## Run

To start all dependencies, run the newly generated docker-compose file:

```bash
docker-compose up
```

Before starting the server, you might want to check your configuration:

```bash
./infomark console configuration test infomark-config.yml
```

This makes InfoMark try to talk to the database, RabbitMQ, and Redis from the
Docker Compose setup. It also tests whether it can save uploads. You will
probably be told that a privacy-statement file does not exist; we ship one
example of a privacy statement in German.

If everything is green, start the server with:

```bash
./infomark serve
```

That's all. The `serve` command takes care of initializing the database the
first time it starts. Point your browser to http://localhost:2020, which
displays the InfoMark login page. To additionally enable two background
workers, run:

```bash
sudo ./infomark work -n 2
```

`sudo` is required to start Docker containers on behalf of the current user
(unless you have added the user to the docker group). The configuration for
the worker and Docker environment can be tested as well:

```bash
./infomark console configuration test infomark-config.yml --test worker
```

Upgrading InfoMark is also easy: stop the InfoMark server and worker, replace
the binary, and start the server and worker again.

## First user

To add a user, register in the web interface. After registration, the email
address needs to be confirmed. If sendmail is configured, you will receive an
instruction email containing a link to activate the account.

Let us activate the user manually using the console. We will also upgrade the
user to have root privileges:

```bash
# confirm email
./infomark console user confirm your@email.com

# find the id of a user
./infomark console user find your@email.com

    1 YourFirstname YourLastname your@email.com

# add the user with id "1" to admins.
./infomark console admin add 1
```

When running InfoMark on a server in production, we recommend using
[NGINX](https://www.nginx.org/) or [Caddy](https://caddyserver.com/) as a
reverse proxy in front of InfoMark.

# Design choices

InfoMark is designed to run within IT-controlled private environments, in
public clouds, or on your own servers, so that you stay compliant with any
data-privacy requirements and retain data sovereignty.

It is based on several design choices:

**Be open, never lock in, and stay easy to extend**<br>
Every part must be open-source, scalable, reliable, and robust. It must be
easy to extract and use the information outside of InfoMark. Development must
be open, and adapting the implementation has to be possible. Writing scripts
(for example, in Python) for common jobs must be easy.

**Be user-friendly to grow**<br>
The entire system must be easy to deploy, maintain, and update, even for
non-technical users with basic IT skills. We want to provide decisions, not
options. Administration should be near zero. You probably have better things
to do than playing server administrator.

**Be robust and reliable to earn trust**<br>
Automatic testing of programming assignments must be language-agnostic,
isolated, and safe. All intensive operations must be scheduled
asynchronously. The frontend must be lightweight, fast, and responsive.
Creating and restoring a database backup should not cause a nervous
breakdown.

**Be modern and simple**<br>
We deliberately chose Go for the backend and React with TypeScript for the
frontend. We had a hard time redeploying our
[old system](https://github.com/infomark-org/InfoMark-deprecated) written in
Ruby on Rails. There should be no magic behind the scenes that breaks when
updating the dependencies.

**Be as general as necessary and not as general as possible**<br>
We deliberately narrowed the feature set down to stay robust. A discussion
board or PDF-annotation support is out of scope by design. InfoMark solves a
very specific problem: the automated testing of homework programming
exercises.

# System overview

This section provides a brief overview of the InfoMark system, including a
description of its parts. We use continuous-integration tests to ensure the
implementation can be built and passes all tests at any point.

At its core, InfoMark is a single compiled Go binary exposed as a RESTful JSON
web server with JavaScript clients. The REST API is described by an OpenAPI
(Swagger) specification. Run `go generate ./...` to regenerate it
(`docs/generate.go` writes `api.yaml`), then open the resulting `api.yaml` in
any Swagger UI to browse the routes.

## Backend

The backend acts as a RESTful JSON web server and is written in
[Go](https://golang.org/). All dependencies are encapsulated in a
docker-compose configuration file. The dependencies are:

- A [PostgreSQL](https://www.postgresql.org/) database to store all dynamic
  data.
- Computationally intensive operations are scheduled and balanced across
  several background workers asynchronously via
  [RabbitMQ](https://www.rabbitmq.com/).
- [Redis](https://redis.io/) as a lightweight key-value memory store.
- [Docker](https://www.docker.com/) as a lightweight sandbox to run
  auto-tests of solutions to programming assignments in an isolated
  environment.

Each exercise task can be linked to a Docker image and a zip file containing
the test code to support testing. See the Administrator's Guide for more
details.

### Server

Part of the backend is the **server**. The server speaks RESTful JSON to the
remote CLI, the web interface, and the workers.

### Workers

The other part of the backend is the **workers**, which are separate
processes that handle the auto-testing of uploads. These workers *can* be
distributed across multiple machines. We recommend one worker process per 100
students. Workers can be added or removed at any time. InfoMark uses AMQP as
its message broker. Each submission is held in a queue, and each worker
executes one job at a time to avoid overloading the system. Our recommendation
is one worker per available CPU core.

The amount of memory used per submission can be configured. Memory swapping is
deactivated.

### Console

To avoid manual interaction with the database, InfoMark provides a console for
running several commands, such as enrolling a student into a course or group
and setting the role of a user.

```bash
./infomark console user find jane.doe

42 Jane Doe jane.doe@student.uni-tuebingen.de

./infomark console user confirm

    Usage:
      infomark console user confirm [email] [flags]

./infomark console user confirm jane.doe@student.uni-tuebingen.de

./infomark console course enroll

    Usage:
      infomark console course enroll [courseID] [userID] [role] [flags]

# roles: student, tutor/ta, admin
./infomark console course enroll 1 42 admin

```

## Frontend

The frontend is a single-page application (SPA) written in React 19 with
TypeScript and the [Ant Design](https://ant.design/) component library. It
lives in the `ui/` directory of this repository and is built with
[Bun](https://bun.sh): `bun run build:embed` type-checks and compiles the app
into the repository-root `static/` directory, which the Go binary bakes in at
compile time via `go:embed`. As a result, the server distributes the static
web interface and the REST API from a single binary; there is no separate UI
release to download. The REST API used by the frontend is described by the
OpenAPI (Swagger) specification generated with `go generate ./...` (see the
top of this section).

# Development

The initial system was developed in the
[computer graphics group](https://uni-tuebingen.de/en/faculties/faculty-of-science/departments/computer-science/lehrstuehle/computergrafik/computer-graphics/)
of the University of Tübingen, because there were no comparable systems that
met our requirements.

### Things that could be different

During the design of the system, we had to make compromises given some
constraints:

* The system has its own user management. Unfortunately, we did not have the
  option to rely on external OAuth2. This is bad (user management is
  difficult) and good (it helps us stay compliant with the GDPR).
* To avoid the hassle of setting up a proper email service, we use a
  pre-installed sendmail configuration. Creating a working email service is
  hard. Creating a proper and secure email service is almost impossible.
