DROP FUNCTION IF EXISTS public.get_timesheet_count;

CREATE
    OR REPLACE FUNCTION public.get_timesheet_count(
    keyword text,
    from_date timestamp with time zone,
    to_date timestamp with time zone,
    location_id_arg text,
    staff_id_arg text
) RETURNS SETOF timesheet_count
    LANGUAGE sql
    STABLE AS
$$
SELECT count(timesheet_id)              AS all_count,
       coalesce(sum(CASE timesheet_status
                        WHEN 'TIMESHEET_STATUS_DRAFT'::text THEN 1
                        ELSE 0 END), 0) AS draft_count,
       coalesce(sum(CASE timesheet_status
                        WHEN 'TIMESHEET_STATUS_SUBMITTED'::text THEN 1
                        ELSE 0 END), 0) AS submitted_count,
       coalesce(sum(CASE timesheet_status
                        WHEN 'TIMESHEET_STATUS_APPROVED'::text THEN 1
                        ELSE 0 END), 0) AS approved_count,
       coalesce(sum(CASE timesheet_status
                        WHEN 'TIMESHEET_STATUS_CONFIRMED'::text THEN 1
                        ELSE 0 END), 0) AS confirmed_count,
       t.resource_path                  AS resource_path
FROM timesheet t
         LEFT JOIN users u ON t.staff_id = u.user_id
WHERE (t.deleted_at IS NULL)
  AND (
    timesheet_date BETWEEN from_date
        AND to_date
    )
  AND t.location_id = COALESCE(location_id_arg, t.location_id)
  AND t.staff_id = COALESCE(staff_id_arg, t.staff_id)
  AND (keyword IS NULL OR u.name ILIKE keyword)
  AND (
        t.timesheet_status = 'TIMESHEET_STATUS_CONFIRMED'::text OR
        (
                    t.timesheet_status = ANY (
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
                                               WHERE tlh.timesheet_id = t.timesheet_id
                                                 AND tlh.flag_on = TRUE
                                                 AND tlh.deleted_at IS NULL
                                           )
                            )
                            OR (
                                SELECT EXISTS(
                                               SELECT 1
                                               FROM other_working_hours owh
                                               WHERE owh.timesheet_id = t.timesheet_id
                                                 AND owh.deleted_at IS NULL
                                           )
                            )
                            OR (
                                SELECT EXISTS(
                                               SELECT 1
                                               FROM transportation_expense te
                                               WHERE te.timesheet_id = t.timesheet_id
                                                 AND te.deleted_at IS NULL
                                           )
                            )
                        )
            )
    )
GROUP BY t.resource_path
$$