#!/usr/bin/env bash

# Exit on errors and on command pipe failures.
set -e

EXPORTER_COMMAND="/usr/bin/medusa_exporter"

# Start Cassandra
cassandra &

# Wait for Cassandra to be ready.
sleep 30
attempt=1
max_attempts=10
while (( attempt <= max_attempts )); do
  if cqlsh -e "SHOW HOST" 2>/dev/null; then
    break
  fi
  sleep 5
  ((attempt++))
done
if (( attempt > max_attempts )); then
  echo "Cassandra did not become ready."
  exit 1
fi

# Create test data.
cqlsh -e "CREATE KEYSPACE IF NOT EXISTS testks WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};"
cqlsh -e "CREATE TABLE IF NOT EXISTS testks.tbl1 (id int PRIMARY KEY, val text);"
cqlsh -e "INSERT INTO testks.tbl1 (id, val) VALUES (1, 'first');"
# Full backup.
medusa backup --mode full --backup-name "full"
# Insert additional data.
cqlsh -e "INSERT INTO testks.tbl1 (id, val) VALUES (2, 'second');"
# Differential backup.
medusa backup --mode differential --backup-name "differential"
# Full backup with prefix.
medusa --prefix demo backup --mode full --backup-name "demo_full"
# Update exporter params.
if [[ -n "${EXPORTER_CONFIG}" ]]; then
	EXPORTER_COMMAND="${EXPORTER_COMMAND} --web.config.file=${EXPORTER_CONFIG}"
fi
# Apply prefix if provided via env var MEDUSA_PREFIX.
if [[ -n "${MEDUSA_PREFIX}" ]]; then
    EXPORTER_COMMAND="${EXPORTER_COMMAND} --medusa.prefix=${MEDUSA_PREFIX}"
fi
# Run medusa_exporter.
exec ${EXPORTER_COMMAND}
