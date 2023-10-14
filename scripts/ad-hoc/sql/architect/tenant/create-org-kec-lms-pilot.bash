#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="mastermgmt"

TENANT_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

INSERT INTO organizations (organization_id, tenant_id,               name,       resource_path, domain_name, logo_url, country,      created_at, updated_at, deleted_at)
VALUES                    ('-2147483627',   '${TENANT_ID}', 'KEC', '-2147483627', 'kec-gr', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/tenant_logo/kec-lms-pilot-logo.png',     'COUNTRY_JP', now(),      now(),      null      );

EOF

