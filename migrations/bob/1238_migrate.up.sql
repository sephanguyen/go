CREATE TABLE IF NOT EXISTS public.notification_student_courses
(
    course_id text COLLATE pg_catalog."default" NOT NULL,
    student_id text COLLATE pg_catalog."default" NOT NULL,
    subscription_id text COLLATE pg_catalog."default" NOT NULL,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamp with time zone,
    resource_path text COLLATE pg_catalog."default" DEFAULT autofillresourcepath(),

	CONSTRAINT pk__notification_student_courses PRIMARY KEY (course_id, student_id, subscription_id)
);

CREATE POLICY rls_notification_student_courses 
ON "notification_student_courses" 
using (permission_check(resource_path, 'notification_student_courses')) 
with check (permission_check(resource_path, 'notification_student_courses'));

ALTER TABLE "notification_student_courses"
    ENABLE ROW LEVEL security;
ALTER TABLE "notification_student_courses"
    FORCE ROW LEVEL security;
