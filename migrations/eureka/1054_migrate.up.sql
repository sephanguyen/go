CREATE TABLE IF NOT EXISTS public.course_student_subscriptions (
    "course_student_subscription_id" text NOT NULL,
    "course_student_id" text NOT NULL,
    "course_id" text NOT NULL,
    "student_id" text NOT NULL,
    "start_at" TIMESTAMP WITH TIME ZONE,
    "end_at" TIMESTAMP WITH TIME ZONE,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" text DEFAULT autofillresourcepath(),

    CONSTRAINT course_students_subscriptions_pk PRIMARY KEY (course_student_subscription_id),
    CONSTRAINT course_students_subscriptions_course_students_fk FOREIGN KEY (course_student_id) REFERENCES "course_students"(course_student_id)
);
/* set RLS */
CREATE POLICY rls_course_student_subscriptions ON "course_student_subscriptions" using (permission_check(resource_path, 'course_student_subscriptions')) with check (permission_check(resource_path, 'course_student_subscriptions'));

ALTER TABLE "course_student_subscriptions" ENABLE ROW LEVEL security;
ALTER TABLE "course_student_subscriptions" FORCE ROW LEVEL security;

INSERT INTO public.course_student_subscriptions (
    course_student_subscription_id,
    course_student_id,
    course_id,
    student_id,
    start_at,
    end_at,
    created_at,
    updated_at,
    deleted_at,
    resource_path
)
SELECT
	generate_ulid() AS course_student_subscription_id,
	course_student_id,
    course_id,
    student_id,
    start_at,
    end_at,
    NOW(),
    NOW(),
    deleted_at,
    resource_path
FROM
	public.course_students cs;