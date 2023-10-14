CREATE TABLE IF NOT EXISTS "lesson_student_subscriptions" (
    "student_subscription_id" TEXT NOT NULL PRIMARY KEY,
    "course_id" TEXT NOT NULL,
    "student_id" TEXT NOT NULL,
    "subscription_id" TEXT NOT NULL,
    "start_at" timestamp with time zone,
    "end_at" timestamp with time zone,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" TEXT,
    CONSTRAINT lesson_student_subscriptions_courses_fk FOREIGN KEY (course_id) REFERENCES "courses"(course_id),
    CONSTRAINT lesson_student_subscriptions_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id)
);