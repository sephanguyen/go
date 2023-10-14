DROP FUNCTION IF EXISTS location_timesheets_non_confirmed_count_v2;

CREATE OR REPLACE FUNCTION public.location_timesheets_non_confirmed_count_v2(
    keyword TEXT,
    from_date timestamptz,
    to_date timestamptz,
    "limit" INT,
    "offset" INT)
    RETURNS SETOF location_timesheet_count AS
$$
SELECT loc.location_id                                      AS location_id,
       name                                                 AS name,
       (SELECT count(timesheet_id)
        FROM timesheet t
        WHERE ((t.location_id = loc.location_id)
            AND (t.timesheet_date BETWEEN from_date AND to_date)
            AND (t.timesheet_status = ANY
                 (ARRAY ['TIMESHEET_STATUS_DRAFT'::text,
                     'TIMESHEET_STATUS_SUBMITTED'::text,
                     'TIMESHEET_STATUS_APPROVED'::text])))) AS count,
       deleted_at,
       resource_path
FROM locations loc
WHERE name ILIKE keyword
ORDER BY count DESC
LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;
