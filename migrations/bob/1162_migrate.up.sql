CREATE TABLE IF NOT EXISTS "lesson_student_subscription_access_path" (
    "student_subscription_id" text NOT NULL,
    "location_id" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" text DEFAULT autofillresourcepath(),
    CONSTRAINT lesson_student_subscription_access_path_lesson_student_subscriptions_fk FOREIGN KEY (student_subscription_id) REFERENCES "lesson_student_subscriptions"(student_subscription_id),
    CONSTRAINT lesson_student_subscription_access_path_locations_fk FOREIGN KEY (location_id) REFERENCES "locations"(location_id),
    CONSTRAINT lesson_student_subscription_access_path_pk PRIMARY KEY (student_subscription_id, location_id)
);

CREATE POLICY rls_lesson_student_subscription_access_path ON "lesson_student_subscription_access_path" using (permission_check(resource_path, 'lesson_student_subscription_access_path')) with check (permission_check(resource_path, 'lesson_student_subscription_access_path'));

ALTER TABLE "lesson_student_subscription_access_path" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_student_subscription_access_path" FORCE ROW LEVEL security;
