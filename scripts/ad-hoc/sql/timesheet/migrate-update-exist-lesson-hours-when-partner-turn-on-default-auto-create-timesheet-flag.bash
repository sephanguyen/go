#!/bin/bash

# This script use to update timesheet lesson hours when turn on auto-create timesheet default for KEC partner
# Example: ./scripts/ad-hoc/sql/timesheet/migrate-update-exist-lesson-hours-when-partner-turn-on-default-auto-create-timesheet-flag.bash

set -euo pipefail

DB_NAME="timesheet"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

  BEGIN;

  SET TIMEZONE = 'Asia/Tokyo';

  UPDATE timesheet_lesson_hours t
	SET flag_on = true, updated_at = NOW()
  FROM lessons l 
  WHERE 
    t.deleted_at IS NULL 
    AND t.timesheet_id IN (
      SELECT timesheet_id 
		    FROM timesheet
		    WHERE 
          deleted_at IS NULL 
			    AND timesheet_date >= now()
			    AND timesheet_status <> 'TIMESHEET_STATUS_APPROVED'
          AND resource_path = '-2147483642'
    )
    AND t.lesson_id = l.lesson_id
		AND l.start_time >= NOW()
    AND t.flag_on <> true
    AND l.resource_path ='-2147483642'; -- KEC

  COMMIT;
EOF

