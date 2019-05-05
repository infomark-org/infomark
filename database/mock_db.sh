python3 mock.py
PGPASSWORD=pass psql -h 'localhost' -p 5433 -U 'user' -d 'db' -f mock.sql > /dev/null