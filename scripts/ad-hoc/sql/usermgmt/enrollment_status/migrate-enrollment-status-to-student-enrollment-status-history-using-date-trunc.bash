set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
INSERT INTO student_enrollment_status_history (student_id, location_id, enrollment_status, start_date, resource_path)
SELECT s.student_id, uap.location_id, s.enrollment_status, date_trunc('second',s.created_at) AS start_date, s.resource_path AS resource_path
FROM students AS s
JOIN users AS u ON s.student_id = u.user_id
JOIN user_access_paths AS uap ON u.user_id = uap.user_id
WHERE s.deleted_at IS NULL
AND uap.deleted_at IS NULL
AND s.resource_path=ANY('{${ORG_ID}}')
AND NOT EXISTS (
    SELECT 1 
    FROM student_enrollment_status_history AS sesh 
    WHERE sesh.student_id = s.student_id 
        AND sesh.location_id = uap.location_id 
        AND sesh.enrollment_status = s.enrollment_status 
        AND sesh.start_date <> s.created_at
)
ON CONFLICT ON CONSTRAINT pk__student_enrollment_status_history DO NOTHING;
EOF
