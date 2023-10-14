#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
WITH org_loc AS (
  SELECT l.*
  FROM locations l
    INNER JOIN location_types lt ON l.location_type = lt.location_type_id
  WHERE l.deleted_at IS NULL AND l.is_archived = false
    AND lt.name = 'org' -- get location as org level
    AND (l.parent_location_id is null or (length(l.parent_location_id) = 0))
    -- AND l.resource_path != '-2147483644'
),
student_were_deleted_accesspath AS (
  SELECT user_id
  FROM user_access_paths uap
  GROUP BY uap.user_id
  HAVING bool_and(deleted_at IS NOT null)
)

INSERT INTO user_access_paths
  (user_id, location_id, created_at, updated_at, resource_path)
SELECT DISTINCT(students.student_id), tem.location_id, now(), now(), students.resource_path
FROM students
  INNER JOIN users
    ON students.student_id = users.user_id
  INNER JOIN (
    SELECT location_id, resource_path FROM org_loc
  ) AS tem ON tem.resource_path = students.resource_path
  LEFT JOIN user_access_paths uap
    ON students.student_id = uap.user_id
  LEFT JOIN (
    SELECT user_id FROM student_were_deleted_accesspath
  ) AS tem2 ON students.student_id = tem2.user_id
WHERE uap.user_id IS NULL
  OR tem2.user_id IS NOT NULL
ON CONFLICT ON CONSTRAINT user_access_paths_pk DO UPDATE SET updated_at = now(), deleted_at = null;
EOF
