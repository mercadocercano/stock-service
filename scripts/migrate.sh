#!/bin/bash

set -e

echo "Running migrations..."

# Asegurarse de que la base de datos stock_db existe
echo "Asegurando que la base de datos $DB_NAME existe..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d postgres -c "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d postgres -c "CREATE DATABASE $DB_NAME"

# Check if any migrations exist
if [ "$(find /migrations -name "*.sql" | wc -l)" -eq 0 ]; then
  echo "No migrations found, skipping."
  exit 0
fi

# Sort migrations by name
MIGRATIONS=$(find /migrations -name "*.sql" | sort)

# Apply each migration
for MIGRATION in $MIGRATIONS; do
  FILENAME=$(basename -- "$MIGRATION")
  echo "Applying migration: $FILENAME"
  
  # Execute the migration script
  PGPASSWORD=$POSTGRES_PASSWORD psql -h $POSTGRES_HOST -U $POSTGRES_USER -d $DB_NAME -f $MIGRATION
  
  # Check for errors
  if [ $? -ne 0 ]; then
    echo "Migration failed: $FILENAME"
    exit 1
  fi
  
  echo "Migration completed: $FILENAME"
done

echo "All migrations completed successfully!" 