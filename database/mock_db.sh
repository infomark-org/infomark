python3 mock.py
PGPASSWORD=postgres psql -h 'localhost' -U 'postgres' -d 'infomark' -f mock.sql > /dev/null