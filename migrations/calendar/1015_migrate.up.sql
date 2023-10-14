
CREATE TABLE IF NOT EXISTS public.time_slot (
	id int4 NOT NULL,
	"year" int4 NULL,
	"period" int4 NULL,
	center_num int4 NULL,
	time_period int4 NULL,
    start_time time NULL,
    end_time time NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT time_slot_pk PRIMARY KEY (id)
);

CREATE POLICY rls_time_slot ON "time_slot" using (permission_check(resource_path, 'time_slot')) with check (permission_check(resource_path, 'time_slot'));
CREATE POLICY rls_time_slot_restrictive ON "time_slot" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'time_slot')) with check (permission_check(resource_path, 'time_slot'));

ALTER TABLE "time_slot" ENABLE ROW LEVEL security;
ALTER TABLE "time_slot" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.job_schedule_status (
	id int4 NOT NULL,
    scheduling_name text NOT NULL,
    start_week int4 NOT NULL,
    end_week int4 NOT NULL,
	location_id text NOT NULL,
    job_date DATE NULL,
    job_time TIME NULL,
    job_status VARCHAR(255),
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamptz NULL,
    resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT job_schedule_status_pk PRIMARY KEY (id)
);

CREATE POLICY rls_job_schedule_status ON "job_schedule_status" USING (permission_check(resource_path, 'job_schedule_status')) WITH CHECK (permission_check(resource_path, 'job_schedule_status'));
CREATE POLICY rls_job_schedule_status_restrictive ON "job_schedule_status" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'job_schedule_status')) WITH CHECK (permission_check(resource_path, 'job_schedule_status'));

ALTER TABLE "job_schedule_status" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "job_schedule_status" FORCE ROW LEVEL SECURITY;

ALTER TABLE public.applied_slot ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.applied_slot ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.applied_slot ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

ALTER TABLE public.center_opening_slot ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.center_opening_slot ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.center_opening_slot ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

ALTER TABLE public.student_available_slot_master ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.student_available_slot_master ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.student_available_slot_master ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

ALTER TABLE public.teacher_subject ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.teacher_subject ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.teacher_subject ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;

ALTER TABLE public.teacher_available_slot_master ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.teacher_available_slot_master ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now());
ALTER TABLE public.teacher_available_slot_master ADD COLUMN IF NOT EXISTS deleted_at timestamptz NULL;
