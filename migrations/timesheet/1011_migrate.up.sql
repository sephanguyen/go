CREATE TABLE IF NOT EXISTS lessons_teachers
(
    lesson_id     TEXT                                                          NOT NULL,
    teacher_id    TEXT                                                          NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    deleted_at    TIMESTAMP WITH TIME ZONE,
    resource_path TEXT                     DEFAULT autofillresourcepath(),
    CONSTRAINT lessons_teachers_pk
        PRIMARY KEY (lesson_id, teacher_id)
);
CREATE POLICY rls_lessons_teachers ON lessons_teachers USING (permission_check(resource_path, 'lessons_teachers'))
WITH CHECK (permission_check(resource_path, 'lessons_teachers'));

ALTER TABLE "lessons_teachers"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "lessons_teachers"
    FORCE ROW LEVEL SECURITY;