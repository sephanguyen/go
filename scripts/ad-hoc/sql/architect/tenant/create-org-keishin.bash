#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

TENANT_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

------ Keishin ------
--- Organization ---
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES   ('-2147483626', '${TENANT_ID}', 'Keishin', '-2147483626', 'manan', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant_logo/keishin-logo.png', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;

--- Update config
UPDATE configuration_key
SET default_value = 'on'
WHERE config_key = 'user.enrollment.update_status_manual'
AND default_value = 'off';

UPDATE internal_configuration_value
SET config_value = 'on'
WHERE config_key = 'user.enrollment.update_status_manual'
AND config_value = 'off'
AND resource_path = '-2147483626';
EOF
