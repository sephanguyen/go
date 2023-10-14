-- Dummy table for Function location_timesheets_non_confirmed_count returning
CREATE TABLE IF NOT EXISTS public.location_timesheet_count
(
    location_id   TEXT,
    name          TEXT,
    "count"       BIGINT,
    deleted_at    timestamptz,
    resource_path TEXT
);

CREATE POLICY rls_location_timesheet_count ON "location_timesheet_count"
    USING (permission_check (resource_path, 'location_timesheet_count'))
    WITH CHECK (permission_check (resource_path, 'location_timesheet_count'));

CREATE POLICY rls_location_timesheet_count_restrictive ON "location_timesheet_count" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'location_timesheet_count'))
WITH CHECK (permission_check(resource_path, 'location_timesheet_count'));

ALTER TABLE "location_timesheet_count" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "location_timesheet_count" FORCE ROW LEVEL SECURITY;
