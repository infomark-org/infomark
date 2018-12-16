# Infomark (GO version)

## Setup

```console
user@server$ sudo -u postgres -- createuser --createdb --pwprompt infomark
user@server$ psql < schema.sql
user@server$ go generate
```

## Refactoring-Motivations

- easier deployment (single binary) + hit run
  - no dependecies hell (only Postgres, Docker for testing) !!!!
- GO-backend with api only
  - unit-tests for backend!
  - cli for some actions submission-upload-tool
    - command line tools for students to upload submission
    - command line tools for changing enrollments and groups, gradings, etc.
  - less computional requirements
  - native gRPC to Docker - Service
- React-frontend (JWT)
