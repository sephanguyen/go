CREATE TYPE week_day AS ENUM ('Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday');

CREATE TABLE IF NOT EXISTS public.working_hour (
    working_hour_id TEXT NOT NULL,
    day week_day NOT NULL,
    opening_time TEXT NOT NULL,
    closing_time TEXT NOT NULL,
    location_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT pk__working_hour PRIMARY KEY (working_hour_id),
    CONSTRAINT unique__working_hour_location_id_day UNIQUE (location_id, day)
);

CREATE POLICY rls_working_hour ON "working_hour"
USING (permission_check(resource_path, 'working_hour')) WITH CHECK (permission_check(resource_path, 'working_hour'));
CREATE POLICY rls_working_hour_restrictive ON "working_hour" AS RESTRICTIVE
USING (permission_check(resource_path, 'working_hour'))WITH CHECK (permission_check(resource_path, 'working_hour'));

ALTER TABLE "working_hour" ENABLE ROW LEVEL security;
ALTER TABLE "working_hour" FORCE ROW LEVEL security;
