-- public.date_type definition

CREATE TABLE IF NOT EXISTS public.date_type (
    date_type_id text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone NULL
);

ALTER TABLE public.date_type ADD CONSTRAINT date_type_pk PRIMARY KEY (date_type_id, resource_path);

CREATE POLICY rls_date_type ON "date_type" using (permission_check(resource_path, 'date_type')) with check (permission_check(resource_path, 'date_type'));
ALTER TABLE "date_type" ENABLE ROW LEVEL security;
ALTER TABLE "date_type" FORCE ROW LEVEL security;

-- public.day_info definition
CREATE TYPE day_info_status AS enum  ('none', 'draft', 'published');

CREATE TABLE public.day_info (
	"date" date NOT NULL,
	location_id text NOT NULL,
	date_type_id text NULL,
	opening_time text NULL,
	status public.day_info_status NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	CONSTRAINT day_info_pk PRIMARY KEY (location_id, date)
);

-- public.day_info foreign keys
ALTER TABLE public.day_info ADD CONSTRAINT day_info_fk FOREIGN KEY (date_type_id,resource_path) REFERENCES public.date_type(date_type_id,resource_path);

CREATE POLICY rls_day_info ON "day_info" using (permission_check(resource_path, 'day_info')) with check (permission_check(resource_path, 'day_info'));
ALTER TABLE "day_info" ENABLE ROW LEVEL security;
ALTER TABLE "day_info" FORCE ROW LEVEL security;

--- Init Day type for KEC
INSERT INTO public.date_type
    (date_type_id, resource_path, created_at, updated_at, deleted_at)
VALUES
    ( 'regular', '-2147483642', now(), now(), NULL),
    ( 'seasonal', '-2147483642', now(), now(), NULL),
    ( 'spare', '-2147483642', now(), now(), NULL),
    ( 'closed', '-2147483642', now(), now(), NULL)
ON CONFLICT DO NOTHING;
