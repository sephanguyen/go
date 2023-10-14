CREATE TYPE frequency AS ENUM ('once', 'weekly');
CREATE TABLE IF NOT EXISTS public.scheduler (
    scheduler_id TEXT NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    freq frequency,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT pk__scheduler PRIMARY KEY (scheduler_id)
);
CREATE POLICY rls_scheduler ON public.scheduler using (permission_check(resource_path, 'scheduler')) with check (permission_check(resource_path, 'scheduler'));

ALTER TABLE public.scheduler ENABLE ROW LEVEL security;
ALTER TABLE public.scheduler FORCE ROW LEVEL security;


ALTER TABLE public.lessons ADD COLUMN IF NOT EXISTS scheduler_id TEXT;
