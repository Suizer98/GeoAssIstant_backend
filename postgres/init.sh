#!/bin/bash
set -e

export PGPASSWORD=postgres123;

# Use the default "postgres" database to create the "mangastore" database
psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "postgres" <<-EOSQL
  CREATE DATABASE geoaistore;
  GRANT ALL PRIVILEGES ON DATABASE geoaistore TO "postgres";
EOSQL
