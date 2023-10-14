CREATE TABLE IF NOT EXISTS public.student_parents (
    student_id TEXT NOT NULL,
    parent_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    relationship TEXT NOT NULL,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT student_parents_pk PRIMARY KEY(student_id, parent_id)
);


CREATE POLICY rls_student_parents ON "student_parents"
USING (permission_check(resource_path, 'student_parents')) WITH CHECK (permission_check(resource_path, 'student_parents'));

CREATE POLICY rls_student_parents_restrictive ON "student_parents" AS RESTRICTIVE
USING (permission_check(resource_path, 'student_parents'))WITH CHECK (permission_check(resource_path, 'student_parents'));

ALTER TABLE "student_parents" ENABLE ROW LEVEL security;
ALTER TABLE "student_parents" FORCE ROW LEVEL security;
