# InfoMark Remote CLI

`infomark-remote` is a small command line client that talks to a running
InfoMark server over its REST API (`/api/v1`). It is useful for
administrative and scripting tasks that are inconvenient through the web
UI, such as scripted submission uploads on behalf of students or quick
account lookups.

The canonical home of the project is
<https://github.com/infomark-org/infomark>.

## What it does

The client is intentionally small and exposes the following commands:

- `ping` — check that a server endpoint is reachable.
- `me` — print the account information of the authenticated user.
- `user find [query]` — search for users by name or email.
- `submission upload [courseID] [taskID] [userID] [filename]` — upload a
  zipped submission on behalf of a student.

## Building

The client lives inside the main InfoMark Go module, so it is built from
the repository root:

```bash
go build -o infomark-remote ./clients/cli
```

This produces an `infomark-remote` binary in the current directory.

## Configuration

The client reads the server endpoint and the credentials from environment
variables. When a required value is missing it prompts for it
interactively, so the variables are optional but convenient for scripting:

- `INFOMARK_URL` — base URL of the server, for example
  `http://localhost:2020` (no trailing `/api/v1`).
- `INFOMARK_EMAIL` — email address used to authenticate.
- `INFOMARK_PASSWORD` — password used to authenticate. When unset, the
  client prompts for it without echoing the input.

Commands that only read public endpoints (such as `ping`) require just the
URL. Commands that access protected endpoints (`me`, `user find`,
`submission upload`) additionally exchange the email and password for a
bearer token against `/api/v1/auth/token`.

## Usage examples

Check that a server is up:

```bash
export INFOMARK_URL="http://localhost:2020"
./infomark-remote ping
```

Show the authenticated account:

```bash
export INFOMARK_URL="http://localhost:2020"
export INFOMARK_EMAIL="test@uni-tuebingen.de"
export INFOMARK_PASSWORD="test"
./infomark-remote me
```

Search for users:

```bash
./infomark-remote user find "muster"
```

Upload a submission on behalf of a student (course 1, task 2, user 42):

```bash
./infomark-remote submission upload 1 2 42 ./solution.zip
```

## Notes

Colored output is emitted only when standard output is attached to a
terminal. Set the `NO_COLOR` environment variable (or use a `dumb`
terminal) to disable it, which is handy when capturing output into files.
