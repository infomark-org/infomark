# INstallation

## Create Database

```bash
apt install postgresql postgresql-contrib
su postgres
psql
\password
\q
```

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

