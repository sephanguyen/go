#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF


WITH RECURSIVE ltypes (location_type_id, level, parent_location_type_id) AS (
    SELECT  location_type_id, 0, display_name
    FROM    location_types
    WHERE   parent_location_type_id is null 
    UNION ALL
    SELECT  p.location_type_id, t0.level + 1, p.display_name
    FROM    location_types p
            INNER JOIN ltypes t0 ON t0.location_type_id = p.parent_location_type_id 
)
 UPDATE location_types
		set "level" = ltypes.level
		from ltypes where ltypes.location_type_id = location_types.location_type_id;

EOF

