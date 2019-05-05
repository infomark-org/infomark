PGPASSWORD=pass psql -h 'localhost' -p 5433 -U 'user' -d 'db' -f schema.sql
PGPASSWORD=pass psql -h 'localhost' -p 5433 -U 'user' -d 'db' -f migrations/0.0.1alpha14.sql
PGPASSWORD=pass psql -h 'localhost' -p 5433 -U 'user' -d 'db' -f migrations/0.0.1alpha21.sql

