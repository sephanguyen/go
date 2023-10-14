#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update course_type set name = '講習' where name = 'Seasonal';
update course_type set name = '通常' where name = 'Regular';

EOF

