CREATE TABLE IF NOT EXISTS public.student_enrollment_status_history (
    student_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    enrollment_status TEXT NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    comment TEXT,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT students_enrollment_status_check
    CHECK ((enrollment_status = ANY
    ('{STUDENT_ENROLLMENT_STATUS_POTENTIAL,
    STUDENT_ENROLLMENT_STATUS_ENROLLED,
    STUDENT_ENROLLMENT_STATUS_TEMPORARY,
    STUDENT_ENROLLMENT_STATUS_NONPOTENTIAL,
    STUDENT_ENROLLMENT_STATUS_WITHDRAWN,
    STUDENT_ENROLLMENT_STATUS_GRADUATED,
    STUDENT_ENROLLMENT_STATUS_LOA}'::text[])
    )),

    CONSTRAINT student_enrollment_status_history__location_id__fk FOREIGN KEY (location_id) REFERENCES public.locations(location_id),
    CONSTRAINT student_enrollment_status_history__student_id__fk FOREIGN KEY (student_id) REFERENCES public.students(student_id),
    UNIQUE(student_id, location_id, enrollment_status, start_date, end_date)
);

CREATE POLICY rls_student_enrollment_status_history ON "student_enrollment_status_history"
USING (permission_check(resource_path, 'student_enrollment_status_history'))
WITH CHECK (permission_check(resource_path, 'student_enrollment_status_history'));

CREATE POLICY rls_student_enrollment_status_history_restrictive ON "student_enrollment_status_history" AS RESTRICTIVE
USING (permission_check(resource_path, 'student_enrollment_status_history'))
WITH CHECK (permission_check(resource_path, 'student_enrollment_status_history'));

ALTER TABLE "student_enrollment_status_history" ENABLE ROW LEVEL security;
ALTER TABLE "student_enrollment_status_history" FORCE ROW LEVEL security;

CREATE INDEX IF NOT EXISTS student_enrollment_status_history__start_date_idx ON public.student_enrollment_status_history(start_date);
CREATE INDEX IF NOT EXISTS student_enrollment_status_history__end_date_idx ON public.student_enrollment_status_history(end_date);