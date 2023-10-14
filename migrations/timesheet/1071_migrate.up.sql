CREATE
    OR REPLACE FUNCTION public.location_timesheets_non_confirmed_count_v3(
    keyword text,
    from_date timestamp with time zone,
    to_date timestamp with time zone
) RETURNS SETOF location_timesheet_count_v2
    LANGUAGE sql
    STABLE AS
$$
SELECT l.location_id   AS location_id,
       l.name          AS name,
       sum(
               CASE
                   timesheet_status
                   WHEN 'TIMESHEET_STATUS_DRAFT' :: text THEN 1
                   ELSE 0
                   END
           )           AS draft_count,
       sum(
               CASE
                   timesheet_status
                   WHEN 'TIMESHEET_STATUS_SUBMITTED' :: text THEN 1
                   ELSE 0
                   END
           )           AS submitted_count,
       sum(
               CASE
                   timesheet_status
                   WHEN 'TIMESHEET_STATUS_APPROVED' :: text THEN 1
                   ELSE 0
                   END
           )           AS approved_count,
       sum(
               CASE
                   timesheet_status
                   WHEN 'TIMESHEET_STATUS_CONFIRMED' :: text THEN 1
                   ELSE 0
                   END
           )           AS confirmed_count,
       sum(
               CASE
                   WHEN timesheet_status <> 'TIMESHEET_STATUS_CONFIRMED' :: text THEN 1
                   ELSE 0
                   END
           )           AS unconfirmed_count,
       (
           SELECT EXISTS(
                          SELECT 1
                          FROM timesheet_confirmation_info
                          WHERE timesheet_confirmation_info.location_id = l.location_id
                            AND (
                                  timesheet_confirmation_info.period_id = (
                                  SELECT id
                                  FROM timesheet_confirmation_period
                                  WHERE from_date BETWEEN start_date
                                            AND end_date
                                  LIMIT 1
                              )
                              )
                      )
       )               AS is_confirmed,
       l.deleted_at    AS deleted_at,
       l.resource_path AS resource_path
FROM (
         SELECT *
         FROM timesheet
         WHERE (deleted_at IS NULL)
           AND (
             timesheet_date BETWEEN from_date
                 AND to_date
             )
           AND (
                 timesheet_status = 'TIMESHEET_STATUS_CONFIRMED' :: text
                 OR (
                             timesheet_status = ANY (
                             ARRAY [
                                 'TIMESHEET_STATUS_SUBMITTED'::text,
                                 'TIMESHEET_STATUS_APPROVED'::text,
                                 'TIMESHEET_STATUS_DRAFT'::text]
                             )
                         AND (
                                     (
                                         SELECT EXISTS(
                                                        SELECT 1
                                                        FROM timesheet_lesson_hours tlh
                                                        WHERE tlh.timesheet_id = timesheet.timesheet_id
                                                          AND tlh.flag_on = TRUE
                                                          AND tlh.deleted_at IS NULL
                                                    )
                                     )
                                     OR (
                                         SELECT EXISTS(
                                                        SELECT 1
                                                        FROM other_working_hours owh
                                                        WHERE owh.timesheet_id = timesheet.timesheet_id
                                                          AND owh.deleted_at IS NULL
                                                    )
                                     )
                                     OR (
                                         SELECT EXISTS(
                                                        SELECT 1
                                                        FROM transportation_expense te
                                                        WHERE te.timesheet_id = timesheet.timesheet_id
                                                          AND te.deleted_at IS NULL
                                                    )
                                     )
                                 )
                     )
             )
     ) t
         RIGHT JOIN locations l ON l.location_id = t.location_id
WHERE (l.deleted_at IS NULL)
  AND (l.name ILIKE keyword)
GROUP BY l.location_id
ORDER BY is_confirmed,
         unconfirmed_count DESC;
$$