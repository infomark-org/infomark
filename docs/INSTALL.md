# Installation

## Database

### Install Postgres Database

```bash
sudo apt install postgresql postgresql-contrib
# switch to user "postgres"
su postgres
# open postgres command prompt
psql
# change password to "postgres"
\password
# quit postgres command prompt
\q
```

### Create Database Credentials

```console
postgres@laburnum:/root$ createuser --interactive
Enter name of role to add: infomark
Shall the new role be a superuser? (y/n) n
Shall the new role be allowed to create databases? (y/n) y
Shall the new role be allowed to create more new roles? (y/n) n

postgres@laburnum:/root$ createdb infomark
postgres@laburnum:/root$ psql -d infomark

infomark=# \conninfo
You are connected to database "infomark" as user "postgres" via socket in "/var/run/postgresql" at port "5432".
```


You might optionally edit the time config for postgres

```bash
sudo nano /etc/postgresql/9.5/main/postgresql.conf  # timezone = 'UTC'
sudo service postgresql restart
```

You should use `pgweb` via

```bash
pgweb --host=localhost --user=postgres --pass=postgres --db=infomark
```

and open `http://localhost:8081/` in your browser.


### Create Database

We need to upload the schema from `database/schema.sql` which contains the structure.

```bash
PGPASSWORD=postgres psql -h 'localhost' -U 'postgres' -d 'infomark' -f schema.sql
```

### Create Mockup

For debugging you might want to use a mockup. Generate a mockup by

```bash
python3 mock.py
PGPASSWORD=postgres psql -h 'localhost' -U 'postgres' -d 'infomark' -f mock.sql
```

## InfoMark - Backend

To build and run infomark-backend type

```bash
# build
go build infomark-backend.go
# run
./infomark-backend serve --config infomark-backend.yml
```

If you changed the credentials of postgres above, you will need to edit `infomark-backend.yml` as well.