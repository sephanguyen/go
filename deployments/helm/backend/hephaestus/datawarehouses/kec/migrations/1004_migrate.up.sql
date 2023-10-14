CREATE SCHEMA IF NOT EXISTS bob;

CREATE TABLE IF NOT EXISTS bob.lessons_teachers_public_info (
    lesson_id text NOT NULL,
    staff_id text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT lessons_teachers_pk PRIMARY KEY (lesson_id, staff_id)
);

CREATE TABLE IF NOT EXISTS bob.lessons_courses_public_info (
    lesson_id text NOT NULL,
    course_id text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT lessons_courses_pk PRIMARY KEY (lesson_id, course_id)
);
