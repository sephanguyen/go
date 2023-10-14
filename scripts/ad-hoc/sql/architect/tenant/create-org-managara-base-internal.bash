#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
------ Managara Base Internal ------
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES ('2147483630', 'withus-base-internal-h7qxl', 'Managara Base Internal', '2147483630', 'managara-base-internal', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant-managara-base-logo.png', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;
EOF
