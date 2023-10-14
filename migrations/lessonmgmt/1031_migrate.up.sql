CREATE TABLE IF NOT EXISTS public.student_enrollment_status_history (
    student_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    enrollment_status TEXT NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    comment TEXT,
    order_id text NULL,
    order_sequence_number integer NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT pk__student_enrollment_status_history PRIMARY KEY (student_id, location_id, enrollment_status, start_date)
);

CREATE POLICY rls_student_enrollment_status_history ON "student_enrollment_status_history"
USING (permission_check(resource_path, 'student_enrollment_status_history'))
WITH CHECK (permission_check(resource_path, 'student_enrollment_status_history'));

CREATE POLICY rls_student_enrollment_status_history_restrictive ON "student_enrollment_status_history" AS RESTRICTIVE
USING (permission_check(resource_path, 'student_enrollment_status_history'))
WITH CHECK (permission_check(resource_path, 'student_enrollment_status_history'));

ALTER TABLE "student_enrollment_status_history" ENABLE ROW LEVEL security;
ALTER TABLE "student_enrollment_status_history" FORCE ROW LEVEL security;
