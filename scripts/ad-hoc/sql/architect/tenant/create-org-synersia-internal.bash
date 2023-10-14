#!/bin/bash

set -euo pipefail

DB_USER=$SA
DB_NAME="mastermgmt"
DB_HOST="localhost"
DB_PORT="5432"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

------ Synersia Internal ------
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES ('2147483646', 'synersia-internal-d0m8j', 'Synersia Internal', '2147483646', 'synersia-internal', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/multi-tenant-logo/synersia-logo.png', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;
EOF
