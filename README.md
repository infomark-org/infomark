# InfoMark

InfoMark is an is a scalable, modern and open-source [rewrite of our](https://github.com/infomark-org/InfoMark-deprecated)
online course management system with auto testing of students submissions using unit tests to ease the task of TAs.

For documentation and more details see [https://infomark.org](https://infomark.org). That page also
includes a [Quickstart Guide](https://infomark.org/guides/overview/).

## Development

### Testing

Run once

```bash
go build -o infomark .
./infomark console configuration create > infomark-test-config.yml
./infomark console configuration create-compose infomark-test-config.yml > docker-compose.yml
# The generated config disables email verification and allows 100
# requests per minute, but the test suite expects CI semantics:
# set authentication.email.verify to true and
# authentication.total_requests_per_minute to 10 in infomark-test-config.yml.

# Test run against an actual database and redis.
sudo docker-compose up -d

# Create the database schema.
export INFOMARK_CONFIG_FILE=`realpath infomark-test-config.yml`
./infomark console database migrate

# We mock some data to test against.
cd migration/mock
pip3 install -r requirements.txt
python3 mock.py
sudo apt install postgresql-client
PGPASSWORD=... psql -h 'localhost' -U 'database_user' -d 'infomark' -f mock.sql >/dev/null
cd ../../
```

Tests can run multiple-times as we rollback all changes to the database.
The redis-backed rate limiter keeps its counters between runs, so flush
redis before each run or the login rate-limit test will produce spurious
429 failures:

```bash
redis-cli flushall  # or: docker exec <redis-container> redis-cli flushall
export INFOMARK_CONFIG_FILE=`realpath infomark-test-config.yml`
go test ./... -cover -v --goblin.timeout 15s -coverprofile coverage.out
```


### Building

The React UI lives in `ui/` and is compiled into the `static/` folder,
which the Go binary embeds at compile time via `go:embed`. Build the UI
first, then the server:

```bash
cd ui
bun install
bun run build:embed   # typechecks, builds, and writes to ../static
cd ..
go build -o infomark .
```

The resulting binary serves the UI at the site root and the API under
`/api/v1`. For UI development use `bun run dev` in `ui/`, which starts a
dev server on port 3000 that proxies `/api` to a locally running backend
on port 2020.
