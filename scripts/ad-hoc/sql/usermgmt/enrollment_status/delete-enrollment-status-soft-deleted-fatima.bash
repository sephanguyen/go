set -euo pipefail

DB_NAME="fatima"

ORG_ID=$1
START_AT=$2
END_AT=$3


psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
DELETE FROM student_enrollment_status_history
WHERE resource_path=ANY('{${ORG_ID}}') 
AND deleted_at::date >= '${START_AT}'
AND deleted_at::date <= '${END_AT}'
EOF
