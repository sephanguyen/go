CREATE
    OR REPLACE FUNCTION public.get_non_confirmed_locations(period_date timestamp with time zone) RETURNS SETOF non_confirmed_locations
    LANGUAGE sql
    STABLE AS
$$
SELECT l.location_id   AS location_id,
       l.deleted_at    AS deleted_at,
       l.resource_path AS resource_path
FROM locations l
WHERE (
              l.deleted_at IS NULL
              AND NOT EXISTS(
                  SELECT 1
                  FROM timesheet_confirmation_info tci
                  WHERE tci.location_id = l.location_id
                    AND tci.deleted_at IS NULL
                    AND (
                          tci.period_id = (
                          SELECT tcp.id
                          FROM timesheet_confirmation_period tcp
                          WHERE period_date BETWEEN tcp.start_date
                              AND tcp.end_date
                            AND tcp.deleted_at IS NULL
                          LIMIT 1
                      )
                      )
              )
          );
$$