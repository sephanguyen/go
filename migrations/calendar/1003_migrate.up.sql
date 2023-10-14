-- CREATE TABLE public.date_type;
CREATE TABLE IF NOT EXISTS public.date_type (
	date_type_id text NOT NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	CONSTRAINT date_type_pk PRIMARY KEY (date_type_id,resource_path)
);

CREATE POLICY rls_date_type ON "date_type" using (permission_check(resource_path, 'date_type')) with check (permission_check(resource_path, 'date_type'));
CREATE POLICY rls_date_type_restrictive ON "date_type" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'date_type')) with check (permission_check(resource_path, 'date_type'));
ALTER TABLE "date_type" ENABLE ROW LEVEL security;
ALTER TABLE "date_type" FORCE ROW LEVEL security;

-- CREATE TABLE public.date_info;
CREATE TYPE date_info_status AS enum  ('none', 'draft', 'published');

CREATE TABLE IF NOT EXISTS public.date_info (
	"date" date NOT NULL,
	location_id text NOT NULL,
	date_type_id text NULL,
	opening_time text NULL,
	status public.date_info_status NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	time_zone text NULL DEFAULT current_setting('TIMEZONE'::text),
	CONSTRAINT date_info_pk PRIMARY KEY (location_id, date)
);

-- public.date_info foreign keys
ALTER TABLE public.date_info ADD CONSTRAINT date_info_fk FOREIGN KEY (date_type_id,resource_path) REFERENCES public.date_type(date_type_id,resource_path);

CREATE POLICY rls_date_info ON "date_info" using (permission_check(resource_path, 'date_info')) with check (permission_check(resource_path, 'date_info'));
CREATE POLICY rls_date_info_restrictive ON "date_info" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'date_info')) with check (permission_check(resource_path, 'date_info'));

ALTER TABLE "date_info" ENABLE ROW LEVEL security;
ALTER TABLE "date_info" FORCE ROW LEVEL security;

-- CREATE TABLE public.scheduler;
CREATE TYPE frequency AS ENUM ('once', 'weekly');

CREATE TABLE IF NOT EXISTS public.scheduler (
	scheduler_id text NOT NULL,
	start_date timestamptz NOT NULL,
	end_date timestamptz NOT NULL,
	freq public.frequency NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT pk__scheduler PRIMARY KEY (scheduler_id)
);

CREATE POLICY rls_scheduler ON public.scheduler using (permission_check(resource_path, 'scheduler')) with check (permission_check(resource_path, 'scheduler'));
CREATE POLICY rls_scheduler_restrictive ON "scheduler" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'scheduler')) with check (permission_check(resource_path, 'scheduler'));

ALTER TABLE public.scheduler ENABLE ROW LEVEL security;
ALTER TABLE public.scheduler FORCE ROW LEVEL security;
