CREATE TABLE IF NOT EXISTS "lesson_student_subscription_access_path" (
    "student_subscription_id" text NOT NULL,
    "location_id" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" text DEFAULT autofillresourcepath(),
    CONSTRAINT lesson_student_subscription_access_path_pk PRIMARY KEY (student_subscription_id, location_id)
);

CREATE POLICY rls_lesson_student_subscription_access_path ON "lesson_student_subscription_access_path" USING (permission_check(resource_path, 'lesson_student_subscription_access_path')) WITH CHECK (permission_check(resource_path, 'lesson_student_subscription_access_path'));
CREATE POLICY rls_lesson_student_subscription_access_path_restrictive ON "lesson_student_subscription_access_path" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lesson_student_subscription_access_path')) with check (permission_check(resource_path, 'lesson_student_subscription_access_path'));

ALTER TABLE "lesson_student_subscription_access_path" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_student_subscription_access_path" FORCE ROW LEVEL security;
