#!/bin/bash

# This script queries schema migration version from database.
# It is a simple query meant to test the adhoc workflow.

set -euo pipefail

# Set your own parameters here
DB_NAME="bob"

# The following command runs psql
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
SELECT * FROM schema_migrations;
EOF
