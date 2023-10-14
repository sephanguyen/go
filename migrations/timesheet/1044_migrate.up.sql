CREATE
OR REPLACE FUNCTION public.location_timesheets_non_confirmed_count_v2(
  keyword text,
  from_date timestamp with time zone,
  to_date timestamp with time zone
) RETURNS SETOF location_timesheet_count LANGUAGE sql STABLE AS $$
SELECT
  l.location_id AS location_id,
  l.name AS name,
  count(t.timesheet_id) AS count,
  l.deleted_at AS deleted_at,
  l.resource_path AS resource_path
FROM
  (
    SELECT
      *
    FROM
      timesheet
    WHERE
      (deleted_at IS NULL)
      AND (
        timesheet_date BETWEEN from_date
        AND to_date
      )
      AND (
        (
          timesheet_status = ANY (
            ARRAY [
                     'TIMESHEET_STATUS_SUBMITTED'::text,
                     'TIMESHEET_STATUS_APPROVED'::text]
          )
          OR (
            timesheet_status = 'TIMESHEET_STATUS_DRAFT' :: text
            AND (
              (
                (
                  SELECT
                    COUNT(*)
                  FROM
                    timesheet_lesson_hours tlh
                  WHERE
                    tlh.timesheet_id = timesheet.timesheet_id
                    AND tlh.flag_on = TRUE
                    AND tlh.deleted_at IS NULL
                ) > 0
              )
              OR (
                (
                  SELECT
                    COUNT(*)
                  FROM
                    other_working_hours owh
                  WHERE
                    owh.timesheet_id = timesheet.timesheet_id
                    AND owh.deleted_at IS NULL
                ) > 0
              )
              OR (
                (
                  SELECT
                    COUNT(*)
                  FROM
                    transportation_expense te
                  WHERE
                    te.timesheet_id = timesheet.timesheet_id
                    AND te.deleted_at IS NULL
                ) > 0
              )
              OR (
                remark IS NOT NULL AND remark <> ''
              )
            )
          )
        )
      )
  ) t
  RIGHT JOIN locations l ON l.location_id = t.location_id
WHERE
  (l.deleted_at IS NULL)
  AND (l.name ILIKE keyword)
GROUP BY
  l.location_id
ORDER BY
  count DESC
$$