# Administrator's guide

# General information

This page gives a brief overview of the underlying system and its
capabilities. It is meant for people who would like to install the system. If
you have been provided with such an instance, please refer to the
[Tutor's Guide](tutor.md) instead.

A single InfoMark server instance can serve multiple courses. The linked
background workers are then shared among these courses.

## Configuration

Configuration is done through a YAML configuration file. For an example
configuration, refer to the
[example](https://github.com/infomark-org/infomark/blob/master/configuration/example.yml).

In general, the defaults should do the job. For all secrets such as passwords,
tokens, or keys, use `openssl rand -base64 32` to generate random,
high-quality secrets, or generate the configuration with the InfoMark console.

Let us discuss some of the important settings.

### Auth_JWT

InfoMark supports two authentication systems:

- JSON Web Tokens (JWT), and
- sessions.

If you use JWTs, ensure they stay valid. There is an access token, which
confirms the identity of the requester and must be attached to the header of
each request. The refresh token serves the purpose of creating a new,
short-lived access token.

Sessions are handled on the server side using cookies. A session stays valid
for 24 hours in total but expires within 20 minutes if no action is taken
(with the default configuration). Note that the frontend frequently sends
requests to the backend even without explicit user interaction, for example
when polling for new test logs from the auto-testing feature. As a result,
these sessions are kept alive.

### Email

For technical reasons, our infrastructure only supports the `sendmail` binary
to deliver emails. If you remove the `sendmail_binary` key from the config,
all emails from the system are printed to the terminal instead. Additionally,
each outgoing email carries a footnote

```
Sent by: FirstName LastName
sent via Infomark
```

when it is composed by a user.

### Directories

We use several directories to store uploads, generated files, and common
files such as a privacy statement. The default values work as long as the
paths exist. We made these paths configurable because these files require
different backup strategies. We suggest estimating the required space
beforehand. From our experience, 1000 submissions amount to about 20 MB. You
can limit the size of each submission file in the server settings under
`server.http.limits`.

### Server settings

To balance the trade-off between too many requests and responsiveness, we
added several strategies to avoid blocking actions. We limit the number of
bytes read from the client during a request. The default of 1 MB is enough for
common JSON requests. You can configure these limits for each kind of request
using the following keys:

- `server.http.limits.max_request_json` to limit any JSON request
- `server.http.limits.max_avatar` to limit the image size for the profile
  avatar
- `server.http.limits.max_submission` to limit the size of homework solutions

There is no limit when uploading slides or extra course material.

### Background workers

Background workers can be turned off with the config setting
`distribute_jobs: false`.

## Roles and permissions

InfoMark has a relatively simple permission system. Permissions are tied to a
request identity (the user). Each user is either a normal user or a
`global admin`. A global admin bypasses all permission checks and effectively
has every permission across all hosted courses.

To upgrade or downgrade a user to `global admin`, use the console:

```bash
# upgrade account
./infomark console admin add [the-user-id]
# downgrade account
./infomark console admin remove [the-user-id]
```

Only a global admin can create new courses.

For finer-grained control, each user enrolled in a course is assigned one of
the following roles:

- *admin*: admin for a course -- not a global admin
- *tutor*: teaching assistant who grades homework and leads exercise groups
- *student*: this is the default role

These roles control which resources a user can access; for example, students
cannot see other students' personal information. Slides and material can be
targeted at a specific role, for example to distribute a sample solution to
the TAs. These roles are nested: an admin has all the permissions a tutor and
a student have, and tutors have additional permissions compared with students.

To upgrade a user to course tutor or course admin, use the console:

```bash
# set permission to student
./infomark console course enroll [course-id] [the-user-id] student
# set permission to tutor
./infomark console course enroll [course-id] [the-user-id] tutor
# set permission to admin
./infomark console course enroll [course-id] [the-user-id] admin
```

Note that this enrolls the user in the course if the user is not already
enrolled. Otherwise, it simply updates the user's role in the course.

## Auto-tests

Please refer to the [Tutor's Guide](tutor.md) for more details on writing and
using auto-tests.

Each task of a programming assignment can be linked to a Docker image and a
zip file containing the test code. Technically, the server process, acting as
a RESTful JSON web server, communicates over AMQP with separate processes
called `workers`. These workers can be started on different machines.

> Please note that, in the current version, these workers have global admin
> privileges. This will change in the future. Hence, only start these workers
> on machines you trust (which is a good idea anyway).

### Conventions

A worker communicates directly with the local Docker API and runs a command
like:

```bash
docker run --rm -it --net="none" \
  -v <STUDENT_UPLOAD.ZIP>:/data/submission.zip:ro \
  -v <TASK_SEPCIFIC_TEST.ZIP>:/data/unittest.zip:ro  \
  <YOUR_DOCKER_IMAGE>
```

Note that each test is isolated from the internet, and the uploaded student
solution `<STUDENT_UPLOAD.ZIP>` and the test framework
`<TASK_SEPCIFIC_TEST.ZIP>` are mounted read-only. The worker captures all
output from stdout that lies between two markers. The Docker output

```stdout
this output here before the marker will be ignored

--- BEGIN --- INFOMARK -- WORKER

[ ok ]     done
[ failed ] done

--- END --- INFOMARK -- WORKER

some other output which will be ignored
```

is captured as

```stdout
[ ok ]     done
[ failed ] done
```

If the Docker child returns an exit code other than 0, a default message is
sent instead. We strongly encourage you to test any new test framework locally
using the `docker run` command above.

The workers download the student submission over HTTP and run the tests
locally. Make sure the Docker image used exists or has already been pulled
from, for example, Docker Hub.

We provide [examples](tutor.md) for testing Java, Python, and C++ programming
assignment solutions.

## Exercise groups

InfoMark supports multiple exercise groups per course. These are usually
weekly meetings where exercises are discussed. The owner of an exercise group
can be any user in the system, regardless of role. However, to be allowed to
grade student submissions, the user needs the course-specific role "tutor".

The system collects *bids* from students. Each bid indicates a preference for
the different dates of the exercise groups (10 means best choice, 1 means the
student would prefer another exercise group). By definition, each student is
enrolled in exactly one group, but tutors can supervise several exercise
groups.

To guarantee the best possible assignment between students and exercise
groups -- maximizing overall happiness -- we solve an integer program using
[Symphony](https://projects.coin-or.org/SYMPHONY). Symphony can reliably solve
problem instances with several exercise groups and thousands of students.

The folder `assignment_solver` contains the Docker image. Alternatively, you
can pull the Docker image `patwie/symphony` from Docker Hub.

To export the bids of students, run:

```bash
./infomark console group dump-bids [courseID] [file] [min_per_group] [max_per_group]
```

Here, `min_per_group` and `max_per_group` constrain the number of allowed
participants in each exercise group. If you use the wrong bounds, the problem
might be infeasible to solve; in that case, change these constraints. An
example is:

```bash
./infomark console group dump-bids 1 mycourse 15 30
```

InfoMark then displays the command to create a solution, for example:

```bash
sudo docker run -v "$PWD":/data -it patwie/symphony  /var/symphony/bin/symphony -F /data/mycourse.mod -D /data/mycourse.dat -f /data/mycourse.par > solution.txt

cat solution.txt
...
assign[u276,g2]    1.0000000000
assign[u277,g9]    1.0000000000
assign[u278,g10]   1.0000000000
assign[u279,g10]   1.0000000000
assign[u280,g6]    1.0000000000
assign[u281,g1]    1.0000000000
assign[u282,g8]    1.0000000000
assign[u283,g5]    1.0000000000
assign[u284,g9]    1.0000000000
assign[u285,g5]    1.0000000000
assign[u286,g3]    1.0000000000
assign[u287,g6]    1.0000000000
assign[u288,g4]    1.0000000000
assign[u289,g9]    1.0000000000
assign[u290,g4]    1.0000000000

...
```

The solution `solution.txt` can be fed directly into InfoMark using:

```bash
./infomark console group import-assignments [courseID] solution.txt
```

Executing this command assigns each student to a group. Once students are
assigned to an exercise group, they can no longer change their preference.
However, any course admin can change the assignment manually.

## Backup and restore

The console features commands to create and restore snapshots of the database.
We use PostgreSQL version 11, and the backup/restore routine requires the
binaries `pg_dump`, `dropdb`, `createdb`, `psql`, and `gunzip`, as it merely
wraps these commands and creates pipes between them.

To install the correct binaries, make sure you run these commands **once**:

```bash
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -
RELEASE=$(lsb_release -cs)
echo "deb http://apt.postgresql.org/pub/repos/apt/ ${RELEASE}"-pgdg main | sudo tee  /etc/apt/sources.list.d/pgdg.list
sudo apt update
sudo apt -y install postgresql-11
which gunzip
```

To create a snapshot, run:

```bash
./infomark console database backup path/to/file.sql.gz
```

To load data from a snapshot, run:

```bash
./infomark console database restore path/to/file.sql.gz
```

## Generated files

The InfoMark backend generates files in an internal cron job. The cron jobs
have a setting `cronjob_intervall_*` that specifies the interval at which
these jobs run. One cron job is `submission_zip`, which zips all submissions
together (per group and task) so that tutors and TAs can download the entire
bundle of submissions.

The job creates a lock file to avoid creating the same zip again and to avoid
race conditions. The interval is set via the setting
`cronjob_intervall_submission_zip`.

These files reside in the directory `generated_files_dir` with the name

```
collection-course%d-sheet%d-task%d-group%d.lock
```

where each `%d` is replaced by the corresponding *id*. To regenerate a
specific file (for example, because you extended the due date of an exercise),
simply remove both files:

```bash
rm collection-course%d-sheet%d-task%d-group%d.lock
rm collection-course%d-sheet%d-task%d-group%d.zip
```

# API

The definition of all available routes to the RESTful backend is described in
an OpenAPI (Swagger) specification. Run `go generate ./...` to regenerate it
(`docs/generate.go` writes `api.yaml`), then open the resulting `api.yaml` in
any Swagger UI to browse the routes.

# Metrics

The metrics for [Prometheus](https://prometheus.io/) are served under
`localhost:<port>/metrics`. We strongly suggest not exposing these metrics when
using a reverse proxy such as NGINX.

To visualize these metrics, we have assembled a
[custom Grafana board](https://github.com/infomark-org/infomark-docs/tree/master/metrics).
