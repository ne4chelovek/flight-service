#!/bin/bash

export MIGRATION_DSN="host=${PG_HOST} port=5432 dbname=${PG_DATABASE_NAME} user=${PG_USER} password=${PG_PASSWORD} sslmode=disable"

echo "Applying migrations with DSN: host=${PG_HOST} dbname=${PG_DATABASE_NAME} user=${PG_USER}"

sleep 6

goose -dir "./migrations" postgres "${MIGRATION_DSN}" up -v

echo "Migrations completed"