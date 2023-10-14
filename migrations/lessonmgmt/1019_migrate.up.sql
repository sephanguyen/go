CREATE TYPE day_info_status AS enum  ('none', 'draft', 'published');

CREATE TABLE IF NOT EXISTS public.day_info (
	"date" date NOT NULL,
	location_id text NOT NULL,
	day_type_id text NULL,
	opening_time text NULL,
	status public.day_info_status NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	time_zone text NULL DEFAULT current_setting('TIMEZONE'::text),
	CONSTRAINT day_info_pk PRIMARY KEY (location_id, date)
);

CREATE POLICY rls_day_info ON "day_info" using (permission_check(resource_path, 'day_info')) with check (permission_check(resource_path, 'day_info'));
CREATE POLICY rls_day_info_restrictive ON "day_info" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'day_info')) with check (permission_check(resource_path, 'day_info'));

ALTER TABLE "day_info" ENABLE ROW LEVEL security;
ALTER TABLE "day_info" FORCE ROW LEVEL security;
