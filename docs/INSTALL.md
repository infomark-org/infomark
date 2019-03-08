# Installation Guide




## Installing Ubuntu Server 18.04 LTS

Install the 64-bit version of Ubuntu Server on each machine that hosts one or more of the components.

**To install Ubuntu Server 18.04:**

1. To install Ubuntu Server 18.04, see the `Ubuntu Installation Guide. <https://help.ubuntu.com/18.04/installation-guide/amd64/index.html>`__

2. After the system is installed, make sure that it's up to date with the most recent security patches. Open a terminal window and issue the following commands:

  ``sudo apt update``

  ``sudo apt upgrade``

Now that the system is up to date, you can start installing the components that make up a InfoMark system.

.. _install-ubuntu-1804-postgresql:

## Installing PostgreSQL Database Server


Install and set up the database for use by the InfoMark server. You can install either PostgreSQL or MySQL.

Assume that the IP address of this server is 10.10.10.1.

**To install PostgreSQL on Ubuntu Server 18.04:**

1. Log in to the server that will host the database and issue the following command:

  ``sudo apt install postgresql postgresql-contrib``

  When the installation is complete, the PostgreSQL server is running, and a Linux user account called *postgres* has been created.

2. Log in to the *postgres* account.

  ``sudo --login --user postgres``

3. Start the PostgreSQL interactive terminal.

  ``psql``

4.  Create the InfoMark database.

  ``postgres=# CREATE DATABASE InfoMark;``

5.  Create the InfoMark user 'infouser'.

  ``postgres=# CREATE USER infouser WITH PASSWORD 'infouser_password';``

  .. note::
    Use a password that is more secure than 'infouser_password'.

6.  Grant the user access to the InfoMark database.

  ``postgres=# GRANT ALL PRIVILEGES ON DATABASE InfoMark to infouser;``

7. Exit the PostgreSQL interactive terminal.

  ``postgre=# \q``

8. Log out of the *postgres* account.

  ``exit``

9. *Optional*: Allow PostgreSQL to listen on all assigned IP Addresses. Open ``/etc/postgresql/10/main/postgresql.conf`` as root in a text editor.

  a. Find the following line:

    ``#listen_addresses = 'localhost'``

  b. Uncomment the line and change ``localhost`` to ``*``:

    ``listen_addresses = '*'``

  c. Restart PostgreSQL for the change to take effect:

    ``sudo systemctl restart postgresql``

10. Modify the file ``pg_hba.conf`` to allow the InfoMark server to communicate with the database.

  **If the InfoMark server and the database are on the same machine**:

    a. Open ``/etc/postgresql/10/main/pg_hba.conf`` as root in a text editor.

    b. Find the following line:

      ``local   all             all                        peer``

    c. Change ``peer`` to ``trust``:

      ``local   all             all                        trust``

  **If the InfoMark server and the database are on different machines**:

    a. Open ``/etc/postgresql/10/main/pg_hba.conf`` as root in a text editor.

    b. Add the following line to the end of the file, where *{InfoMark-server-IP}* is the IP address of the machine that contains the InfoMark server.

      ``host all all {InfoMark-server-IP}/32 md5``

11. Reload PostgreSQL:

  ``sudo systemctl reload postgresql``

12. Verify that you can connect with the user *infouser*.

  a. If the InfoMark server and the database are on the same machine, use the following command:

    ``psql --dbname=InfoMark --username=infouser --password``

  b. If the InfoMark server is on a different machine, log into that machine and use the following command:

    ``psql --host={postgres-server-IP} --dbname=InfoMark --username=infouser --password``

    .. note::
      You might have to install the PostgreSQL client software to use the command.

  The PostgreSQL interactive terminal starts. To exit the PostgreSQL interactive terminal, type ``\q`` and press **Enter**.

With the database installed and the initial setup complete, you can now install the InfoMark server.


### Generating Schema and Mocking Data

You should use `pgweb` via

```bash
pgweb --host=localhost --user=infouser --pass=infouser_password --db=InfoMark
```

and open `http://localhost:8081/` in your browser, whenever you want to debug the database.


#### Create Database

We need to upload the schema from `database/schema.sql` which contains the structure.

```bash
PGPASSWORD=infouser_password psql -h 'localhost' -U 'infouser' -d 'InfoMark' -f schema.sql >/dev/null
```

#### Create Mockup

For debugging you might want to use a mockup. Generate a mockup by

```bash
python3 mock.py
PGPASSWORD=infouser_password psql -h 'localhost' -U 'infouser' -d 'InfoMark' -f mock.sql >/dev/null
```



## Building InfoMark - Backend

To build and run infomark-backend type

```bash
# build
go build infomark-backend.go
# edit config
cp infomark-backend.yml.example ~/infomark-backend.yml
edit ~/infomark-backend.yml
# run
./infomark-backend serve
```

## Generating the documentation

The command

```bash
go generate
```

will generate a valid `api.yaml` for Swagger 3.0.
Hereby, it verifies all implemented routes are documented and have the correct method (get, post, patch, put).
Further, it uses the request and response go-structs to generate request and response body information in swagger.