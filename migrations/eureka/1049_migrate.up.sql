CREATE OR REPLACE function permission_check(resource_path TEXT, table_name TEXT)
RETURNS BOOLEAN 
AS $$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$$  LANGUAGE SQL IMMUTABLE;

WITH TMP AS (
    SELECT course_student_id FROM
    (
        SELECT course_student_id,COUNT(*) as total FROM course_students
        GROUP BY course_student_id
    ) cs
    JOIN course_students
    USING(course_student_id)
    WHERE total > 1
) 
UPDATE course_students SET course_student_id = generate_ulid() WHERE course_student_id IN(SELECT * FROM TMP);

ALTER TABLE public.course_students ADD CONSTRAINT course_student_id_un UNIQUE (course_student_id);

CREATE TABLE IF NOT EXISTS public.course_students_access_paths (
    "course_student_id" text NOT NULL,
    "location_id" text NOT NULL,
    "course_id" text NOT NULL,
    "student_id" text NOT NULL,
    "access_path" text,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" text DEFAULT autofillresourcepath(),

    CONSTRAINT course_students_access_paths_pk PRIMARY KEY (course_student_id, location_id),
    CONSTRAINT course_students_access_paths_course_students_fk FOREIGN KEY (course_student_id) REFERENCES "course_students"(course_student_id)
);

CREATE POLICY rls_course_students_access_paths ON "course_students_access_paths" using (permission_check(resource_path, 'course_students_access_paths')) with check (permission_check(resource_path, 'course_students_access_paths'));

ALTER TABLE "course_students_access_paths" ENABLE ROW LEVEL security;
ALTER TABLE "course_students_access_paths" FORCE ROW LEVEL security;
