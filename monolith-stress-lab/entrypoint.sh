#!/bin/bash
set -e

# Define directories
export PGDATA=/var/lib/postgresql/data
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres # Dummy strict requirement for init

chown -R postgres:postgres /var/lib/postgresql /var/run/postgresql

echo "Initializing PostgreSQL database..."
# Initialize DB cluster as postgres user (Alpine postgresql initdb path)
su-exec postgres initdb -D $PGDATA

echo "listen_addresses='localhost'" >> $PGDATA/postgresql.conf

# Start postgres in background for setup
echo "Starting PostgreSQL temporarily..."
su-exec postgres pg_ctl -D $PGDATA -w start

echo "Running DB setup commands..."
# Create our stresslab user and database
su-exec postgres psql -c "CREATE USER stressuser WITH PASSWORD 'stresspassword';" || true
su-exec postgres psql -c "CREATE DATABASE stresslab OWNER stressuser;" || true
su-exec postgres psql -c "ALTER ROLE stressuser SET client_encoding TO 'utf8';" || true

if [ -f "/app/init.sql" ]; then
    echo "Running initialization SQL script..."
    su-exec postgres psql -d stresslab -U stressuser -f /app/init.sql
    echo "Initialization complete."
fi

# Stop postgres temporary instance
su-exec postgres pg_ctl -D $PGDATA -m fast -w stop

# Start PostgreSQL natively
echo "Starting PostgreSQL service..."
su-exec postgres postgres -D $PGDATA &

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be fully ready..."
until su-exec postgres psql -c '\q' 2>/dev/null; do
  sleep 1
done

echo "Starting Go Application..."
exec /app/main
