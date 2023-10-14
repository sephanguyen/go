CREATE TABLE IF NOT EXISTS public.notification_students (
    student_id TEXT NOT NULL,
    current_grade SMALLINT,
    deleted_at timestamp with time zone NOT NULL,
    resource_path text NULL DEFAULT autofillresourcepath(),
    CONSTRAINT notification_students_student_id_un UNIQUE (student_id)
);

CREATE TABLE IF NOT EXISTS public.notification_users (
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    full_name_phonetic TEXT,
    deleted_at timestamp with time zone NOT NULL,
    resource_path text NULL DEFAULT autofillresourcepath(),
    CONSTRAINT notification_users_user_id_un UNIQUE (user_id)
);

CREATE TABLE IF NOT EXISTS public.notification_student_parents (
    student_id TEXT NOT NULL,
    parent_id TEXT NOT NULL,
    deleted_at timestamp with time zone NOT NULL,
    resource_path text NULL DEFAULT autofillresourcepath(),
    CONSTRAINT notification_student_parents_stu_par_id_un UNIQUE (student_id, parent_id)
);

CREATE POLICY rls_notification_students ON "notification_students"
USING (permission_check(resource_path, 'notification_students'))
WITH CHECK (permission_check(resource_path, 'notification_students'));

ALTER TABLE "notification_students" ENABLE ROW LEVEL security;
ALTER TABLE "notification_students" FORCE ROW LEVEL security;

CREATE POLICY rls_notification_users ON "notification_users"
USING (permission_check(resource_path, 'notification_users'))
WITH CHECK (permission_check(resource_path, 'notification_users'));

ALTER TABLE "notification_users" ENABLE ROW LEVEL security;
ALTER TABLE "notification_users" FORCE ROW LEVEL security;

CREATE POLICY rls_notification_student_parents ON "notification_student_parents"
USING (permission_check(resource_path, 'notification_student_parents'))
WITH CHECK (permission_check(resource_path, 'notification_student_parents'));

ALTER TABLE "notification_student_parents" ENABLE ROW LEVEL security;
ALTER TABLE "notification_student_parents" FORCE ROW LEVEL security;