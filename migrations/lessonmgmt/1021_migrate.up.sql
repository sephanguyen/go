CREATE TABLE  IF NOT EXISTS lessons_courses
(
    lesson_id     TEXT                                                          NOT NULL,
    course_id     TEXT                                                          NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    deleted_at    TIMESTAMP WITH TIME ZONE,
    resource_path TEXT                     DEFAULT autofillresourcepath(),
    CONSTRAINT lessons_courses_pk
        PRIMARY KEY (lesson_id, course_id)
);
CREATE POLICY rls_lessons_courses ON lessons_courses USING (permission_check(resource_path, 'lessons_courses'))
WITH CHECK (permission_check(resource_path, 'lessons_courses'));
CREATE POLICY rls_lessons_courses_restrictive ON lessons_courses AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lessons_courses')) with check (permission_check(resource_path, 'lessons_courses'));

ALTER TABLE "lessons_courses"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "lessons_courses"
    FORCE ROW LEVEL SECURITY;