#!/bin/sh

set -e

host="$1"
shift
cmd="$@"

# Intentar conectarse a postgres (base de datos por defecto) en lugar de a stock_db
until PGPASSWORD=$POSTGRES_PASSWORD psql -h "$host" -U "$POSTGRES_USER" -d "postgres" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing command"
exec $cmd 