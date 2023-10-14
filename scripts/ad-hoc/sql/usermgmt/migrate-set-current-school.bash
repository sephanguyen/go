#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_IDs=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
BEGIN;

UPDATE public.school_history sh
SET is_current = FALSE
WHERE sh.resource_path = ANY('{${ORG_IDs}}');

WITH cte AS (
    SELECT sh.* FROM school_history sh
    INNER JOIN school_info si ON sh.school_id = si.school_id
    INNER JOIN school_level sl ON si.school_level_id = sl.school_level_id
    INNER JOIN school_level_grade slg ON slg.school_level_id = sl.school_level_id
    INNER JOIN students s ON s.grade_id = slg.grade_id
    WHERE sh.resource_path = ANY('{${ORG_IDs}}') AND sh.deleted_at IS NULL
)
UPDATE public.school_history sh
SET is_current = TRUE
FROM cte
WHERE sh.school_id = cte.school_id;

COMMIT;
EOF
