#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1
DATE=$2
LOCATIONS=$3

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE public.students s
SET previous_grade = s.current_grade, current_grade = g1.sequence, grade_id = g1.grade_id
FROM public.grade g1, public.user_access_paths uap
WHERE s.deleted_at IS NULL
AND uap.location_id = ANY('{${LOCATIONS}}') 
AND uap.user_id = s.student_id 
AND g1.sequence = LEAST(
    (SELECT g2.sequence+1 FROM public.grade g2 WHERE g2.grade_id = s.grade_id LIMIT 1), 
    (SELECT MAX(g3.sequence) FROM public.grade g3)
)
AND s.created_at::date <= '${DATE}'
AND s.resource_path = '${ORG_ID}';
EOF
