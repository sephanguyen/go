-- Dummy table for Function get_non_confirmed_locations returning
CREATE TABLE IF NOT EXISTS public.non_confirmed_locations
(
    location_id   TEXT,
    deleted_at    timestamptz,
    resource_path TEXT
);

CREATE POLICY rls_non_confirmed_locations ON "non_confirmed_locations"
    USING (permission_check(resource_path, 'non_confirmed_locations'))
    WITH CHECK (permission_check(resource_path, 'non_confirmed_locations'));

CREATE POLICY rls_non_confirmed_locations_restrictive ON "non_confirmed_locations"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'non_confirmed_locations'))
    WITH CHECK (permission_check(resource_path, 'non_confirmed_locations'));

ALTER TABLE "non_confirmed_locations"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "non_confirmed_locations"
    FORCE ROW LEVEL SECURITY;

CREATE
    OR REPLACE FUNCTION public.get_non_confirmed_locations(
    period_date timestamp with time zone
) RETURNS SETOF non_confirmed_locations
    LANGUAGE sql
    STABLE AS
$$
SELECT l.location_id   AS location_id,
       l.deleted_at    AS deleted_at,
       l.resource_path AS resource_path
FROM locations l
WHERE (l.deleted_at IS NULL
    AND l.is_archived = FALSE
    AND
       NOT EXISTS (SELECT 1
        FROM timesheet_confirmation_info tci
        WHERE tci.location_id = l.location_id
          AND tci.deleted_at IS NULL
          AND (tci.period_id = (SELECT period_id
                                FROM timesheet_confirmation_period tcp
                                WHERE period_date BETWEEN tcp.start_date AND tcp.end_date
                                  AND tcp.deleted_at IS NULL
                                LIMIT 1))));
$$