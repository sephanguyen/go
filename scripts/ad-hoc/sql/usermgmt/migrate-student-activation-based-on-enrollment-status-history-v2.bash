#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
WITH sub_query AS (
    SELECT s.student_id, inactive_students.latest_start_date FROM public.students s LEFT JOIN (
        SELECT sesh1.student_id, MAX(sesh1.start_date) AS latest_start_date 
        FROM student_enrollment_status_history sesh1
        WHERE NOT EXISTS 
            (
                SELECT 1
                FROM student_enrollment_status_history sesh2
                WHERE NOT (sesh2.enrollment_status = ANY('{
                    STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL,
                    STUDENT_ENROLLMENT_STATUS_WITHDRAWN,
                    STUDENT_ENROLLMENT_STATUS_GRADUATED
                }'))
                AND sesh2.student_id = sesh1.student_id 
                AND (end_date > NOW() OR end_date IS NULL)
                AND start_date < NOW()
                AND deleted_at is NULL
                AND resource_path = ANY('{${ORG_ID}}')
            )
        AND (end_date > NOW() OR end_date IS NULL)
        AND start_date < NOW()
        AND resource_path = ANY('{${ORG_ID}}')
        AND deleted_at is NULL
        GROUP BY sesh1.student_id
  ) inactive_students ON s.student_id = inactive_students.student_id
)
UPDATE users SET deactivated_at = sub_query.latest_start_date FROM sub_query 
WHERE sub_query.student_id = users.user_id AND users.deleted_at IS NULL AND resource_path = ANY('{${ORG_ID}}')
EOF

# ROLL BACK
###
# WITH sub_query AS (
#     SELECT sesh1.student_id, MAX(sesh1.start_date)
# 	FROM student_enrollment_status_history sesh1
# 	WHERE NOT EXISTS 
# 		(
# 			SELECT 1
# 			FROM student_enrollment_status_history sesh2
# 			WHERE NOT (sesh2.enrollment_status = ANY('{
# 				STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL,
# 				STUDENT_ENROLLMENT_STATUS_WITHDRAWN,
# 				STUDENT_ENROLLMENT_STATUS_GRADUATED
# 			}'))
# 			AND sesh2.student_id = sesh1.student_id 
# 			AND (end_date > NOW() OR end_date IS NULL)
# 			AND start_date < NOW()
#             AND resource_path = ANY('{${ORG_ID}}')
# 		)
# 	AND (end_date > NOW() OR end_date IS NULL)
# 	AND start_date < NOW()
#     AND resource_path = ANY('{${ORG_ID}}')
# 	GROUP BY sesh1.student_id
# )
# UPDATE users SET deactivated_at = NULL FROM sub_query 
# WHERE sub_query.student_id = users.user_id AND users.deleted_at IS NULL AND resource_path = ANY('{${ORG_ID}}')