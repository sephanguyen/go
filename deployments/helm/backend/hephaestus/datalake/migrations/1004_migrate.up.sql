CREATE SCHEMA IF NOT EXISTS bob;

CREATE TABLE IF NOT EXISTS bob.lessons_teachers (
    lesson_id text NOT NULL,
    teacher_id text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL,
	teacher_name text NULL,
    CONSTRAINT lessons_teachers_pk PRIMARY KEY (lesson_id, teacher_id)
);

CREATE TABLE IF NOT EXISTS bob.lessons_courses (
    lesson_id text NOT NULL,
    course_id text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL,
    CONSTRAINT lessons_courses_pk PRIMARY KEY (lesson_id, course_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.lessons_teachers;

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.lessons_courses;

