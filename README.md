# InfoMark

InfoMark is an is a scalable, modern and open-source [rewrite of our](https://github.com/infomark-org/InfoMark-deprecated)
online course management system with auto testing of students submissions using unit tests to ease the task of TAs.

For documentation and more details see [https://infomark.org](https://infomark.org). That page also
includes a [Quickstart Guide](https://infomark.org/guides/overview/).

## Development

### Testing

Run once

```bash
./infomark console configuration create > infomark-test-config.yml
./infomark console configuration create-compose infomark-config.yml > docker-compose.yml
# Test run against an actual database and redis.
sudo docker-compose up

# We mock some data to test against.
cd migration/mock
pip3 install -r requirements.txt
python3 mock.py
sudo apt install postgresql-client
PGPASSWORD=... psql -h 'localhost' -U 'database_user' -d 'infomark' -f mock.sql >/dev/null
cd ../../
```

Tests can run multiple-times as we rollback all changes to the database.

```bash
export INFOMARK_CONFIG_FILE=`realpath infomark-test-config.yml`
go test ./... -cover -v --goblin.timeout 15s -coverprofile coverage.out
```


### Building

You will either need to download the [UI](https://github.com/infomark-org/infomark-ui)
from the release page or build it yourself. The ui has to be copied to the static
folder. ([details](https://github.com/infomark-org/infomark/blob/master/.drone.yml#L85-L101))
