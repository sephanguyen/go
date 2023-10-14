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
