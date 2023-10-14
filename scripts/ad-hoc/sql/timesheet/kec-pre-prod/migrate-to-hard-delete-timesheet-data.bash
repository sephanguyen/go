set -euo pipefail

DB_NAME="timesheet"

ORG_ID=$1


psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
DELETE FROM timesheet_lesson_hours
WHERE resource_path=ANY('{${ORG_ID}}');

DELETE FROM other_working_hours
WHERE resource_path=ANY('{${ORG_ID}}');

DELETE FROM transportation_expense
WHERE resource_path=ANY('{${ORG_ID}}');

DELETE FROM timesheet_action_log
WHERE resource_path=ANY('{${ORG_ID}}');

DELETE FROM timesheet
WHERE resource_path=ANY('{${ORG_ID}}')

EOF
