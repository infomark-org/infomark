PGPASSWORD=pass psql -h '127.0.0.1' -p 5433 -U 'postgres' -d 'db' -f schema.sql
# ./infomark-backend console database migrate