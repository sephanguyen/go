CREATE TABLE IF NOT EXISTS public.students (
	student_id text NOT NULL,
	current_grade int2 NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	student_external_id text NULL,
	grade_id text NULL,
	CONSTRAINT students_pk PRIMARY KEY (student_id)
);

CREATE POLICY rls_students ON "students" USING (permission_check(resource_path, 'students')) WITH CHECK (permission_check(resource_path, 'students'));
CREATE POLICY rls_students_restrictive ON "students" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'students')) with check (permission_check(resource_path, 'students'));

ALTER TABLE "students" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "students" FORCE ROW LEVEL SECURITY;