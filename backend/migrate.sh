#!/bin/sh
set -e

DB_STRING="host=$POSTGRES_HOST port=$POSTGRES_PORT user=$POSTGRES_USER password=$POSTGRES_PASSWORD dbname=$POSTGRES_DB sslmode=$POSTGRES_SSLMODE"

exec goose postgres "$DB_STRING" "$@"
