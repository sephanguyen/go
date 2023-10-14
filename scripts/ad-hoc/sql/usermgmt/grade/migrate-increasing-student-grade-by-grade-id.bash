#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1
USER_IDS=$2
GRADE_ID=$3

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
WITH cte AS (
    SELECT s.student_id, g2.grade_id current_grade, g2."sequence" current_sequence, g1.grade_id new_grade, g1."sequence" new_sequence
    FROM grade g1, students s
	JOIN grade g2 ON s.grade_id = g2.grade_id 
	WHERE g1.grade_id = '${GRADE_ID}'
	AND s.deleted_at IS NULL 
	AND s.student_id = ANY('{${USER_IDS}}') 
    AND s.resource_path = '${ORG_ID}';
)
UPDATE public.students s
SET previous_grade = current_sequence, grade_id = new_grade
FROM cte
WHERE s.student_id = cte.student_id
EOF

### Rollback script ###
# UPDATE public.students s
# SET grade_id = g.grade_id 
# FROM grade g
# WHERE s.previous_grade = g."sequence" 
# and s.student_id = ANY('{${USER_IDS}}')
# AND s.resource_path = '${ORG_ID}';
