DROP FUNCTION IF EXISTS location_timesheets_non_confirmed_count_v2;

CREATE OR REPLACE FUNCTION location_timesheets_non_confirmed_count_v2(
    keyword TEXT,
    from_date timestamptz,
    to_date timestamptz)
    RETURNS SETOF location_timesheet_count AS
$$
SELECT l.location_id         AS location_id,
       l.name                AS name,
       count(t.timesheet_id) AS count,
       l.deleted_at          AS deleted_at,
       l.resource_path       AS resource_path
FROM (
         SELECT *
         FROM timesheet
         WHERE (deleted_at IS NULL)
           AND (timesheet_date BETWEEN from_date
             AND to_date)
           AND (timesheet_status = ANY
                (ARRAY ['TIMESHEET_STATUS_DRAFT'::text,
                        'TIMESHEET_STATUS_SUBMITTED'::text,
                        'TIMESHEET_STATUS_APPROVED'::text]))
     ) t
         RIGHT JOIN locations l ON l.location_id = t.location_id
WHERE (l.deleted_at IS NULL)
  AND (l.name ILIKE keyword)
GROUP BY l.location_id
ORDER BY count DESC
$$ LANGUAGE sql STABLE;
