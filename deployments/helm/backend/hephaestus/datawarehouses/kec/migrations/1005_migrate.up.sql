CREATE SCHEMA IF NOT EXISTS bob;

CREATE TABLE IF NOT EXISTS bob.reallocation_public_info (
    student_id TEXT NULL,
    original_lesson_id TEXT NOT NULL,
    new_lesson_id TEXT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    course_id TEXT NOT NULL,
    CONSTRAINT pk__reallocation PRIMARY KEY (original_lesson_id, student_id)
);

ALTER TABLE bob.scheduler_public_info DROP COLUMN IF EXISTS test;
