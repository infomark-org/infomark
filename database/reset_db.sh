PGPASSWORD=postgres psql -h 'localhost' -U 'postgres' -d 'infomark' -f schema.sql
PGPASSWORD=postgres psql -h 'localhost' -U 'postgres' -d 'infomark' -f migration/0.0.1alpha14.sql

