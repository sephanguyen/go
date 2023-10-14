set -euo pipefail

DB_NAME="bob"

ORG_ID=$1


psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
DELETE FROM student_enrollment_status_history
WHERE resource_path=ANY('{${ORG_ID}}')
EOF
