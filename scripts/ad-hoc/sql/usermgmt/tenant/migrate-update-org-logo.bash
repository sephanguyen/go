#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1
LOGO_URL=$2

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

UPDATE public.organizations
SET logo_url = '${LOGO_URL}'
WHERE organization_id = '${ORG_ID}';

EOF
