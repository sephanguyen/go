-- public.user_access_paths definition

-- Drop table

-- DROP TABLE public.user_access_paths;

CREATE TABLE IF NOT EXISTS public.user_access_paths (
	user_id text NOT NULL,
	location_id text NOT NULL,
	access_path text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT user_access_paths_pk PRIMARY KEY (user_id, location_id)
);
CREATE INDEX IF NOT EXISTS user_access_paths__location_id__idx ON public.user_access_paths USING btree (location_id);
CREATE INDEX IF NOT EXISTS user_access_paths__user_id__idx ON public.user_access_paths USING btree (user_id);

CREATE POLICY rls_user_access_paths ON "user_access_paths" USING (permission_check(resource_path, 'user_access_paths')) WITH CHECK (permission_check(resource_path, 'user_access_paths'));
ALTER TABLE IF EXISTS "user_access_paths"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS "user_access_paths"
    FORCE ROW LEVEL SECURITY;

-- public.course_access_paths definition

-- Drop table

-- DROP TABLE public.course_access_paths;

CREATE TABLE IF NOT EXISTS public.course_access_paths (
	course_id text NOT NULL,
	location_id text NOT NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT course_access_paths_pk PRIMARY KEY (course_id, location_id)
);

CREATE POLICY rls_course_access_paths ON "course_access_paths" USING (permission_check(resource_path, 'course_access_paths')) WITH CHECK (permission_check(resource_path, 'course_access_paths'));
ALTER TABLE IF EXISTS "user_access_paths"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS "user_access_paths"
    FORCE ROW LEVEL SECURITY;
