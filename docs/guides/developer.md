# Developer's guide

The system lives in a single git repository containing both the Go backend and
the React web interface. Parts are tested using continuous integration. It is
a good idea to read the CI configuration to see how the pieces are built from
source.

Whenever a combination is stable, we create a new release, which can be
downloaded from the
[release page](https://github.com/infomark-org/infomark/releases). Each
release is a single binary with the compiled web interface embedded, so there
is nothing else to download.

# Frontend

The frontend is a single-page application written in React 19 with TypeScript
and the [Ant Design](https://ant.design/) component library. It lives in the
`ui/` directory of this repository and is built with [Bun](https://bun.sh).

```bash
cd ui
bun install
bun run build:embed
```

`bun run build:embed` type-checks and compiles the app into the
repository-root `static/` directory, which the Go binary embeds at compile
time via `go:embed`. For frontend development, run `bun run dev` in `ui/`,
which starts a dev server that proxies API requests to a locally running
backend.

# Backend

The backend is written in [Go](https://golang.org/). Build the UI first (see
above), then compile the server so the compiled UI is embedded into the
binary:

```bash
git clone https://github.com/infomark-org/infomark
cd infomark
go build -o infomark .
```

Building InfoMark requires Go version 1.13 or newer.

## Unit tests

To guarantee a stable version, each commit is tested against a set of unit
tests.

> While most projects test against a mock, we test the actual behavior of the
> endpoints against a real PostgreSQL database.

To run the tests, set up the dependencies as described in the
[Overview](overview.md). Make sure the database is empty but migrated to the
latest version. You can achieve this by dropping all data:

```bash
sudo docker-compose down -v
sudo docker-compose up
./infomark console database migrate
```

In addition, we add mock entries to the database.

```bash
# mock database content
cd migration/mock
python3 mock.py

PGPASSWORD=pass psql -h YOUR_HOST -U YOUR_USER -p YOUR_PORT -d YOUR_DBNAME -f mock.sql
```

Then

```bash
go test ./... --cover
```

runs the unit tests. Each test is handled in a transaction and does not change
the database (using commit and rollback). However, some test cases depend on
the actual `mock.sql` data.

## Features and scope

InfoMark aims to be as generic as necessary, not as generic as possible. The
following features are intentionally not implemented:

* in-browser PDF annotation
* discussion board
* email conversation and inbox
