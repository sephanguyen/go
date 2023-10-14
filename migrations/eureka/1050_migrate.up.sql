CREATE TABLE IF NOT EXISTS public.study_plan_monitors (
    study_plan_monitor_id TEXT,
    student_id TEXT,
    course_id TEXT,
    type TEXT,
    payload JSONB,
    level TEXT, -- consider next time, such like: critical, high, medium, low
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT study_plan_monitor_pk PRIMARY KEY (study_plan_monitor_id)
);

/* set RLS */
CREATE POLICY rls_study_plan_monitors ON "study_plan_monitors" using (permission_check(resource_path, 'study_plan_monitors')) with check (permission_check(resource_path, 'study_plan_monitors'));

ALTER TABLE "study_plan_monitors" ENABLE ROW LEVEL security;
ALTER TABLE "study_plan_monitors" FORCE ROW LEVEL security;