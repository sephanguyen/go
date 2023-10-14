CREATE SCHEMA IF NOT EXISTS bob;

CREATE TABLE IF NOT EXISTS bob.reallocation (
    student_id TEXT NULL,
    original_lesson_id TEXT NOT NULL,
    new_lesson_id TEXT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    course_id TEXT NOT NULL,
    resource_path TEXT NOT NULL,
    CONSTRAINT pk__reallocation PRIMARY KEY (original_lesson_id, student_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.reallocation;
