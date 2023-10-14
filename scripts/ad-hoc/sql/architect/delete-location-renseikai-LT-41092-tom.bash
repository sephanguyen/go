#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="tom"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update conversation_locations set deleted_at = '2023-04-04 04:04:44.444 +0700' 
where location_id in ('01GP3BN89ES8VHECCVRCHDDV64', '01GP3CMSRY7T288H60GEGDVEWS') and deleted_at is null and resource_path = '-2147483645';

EOF

