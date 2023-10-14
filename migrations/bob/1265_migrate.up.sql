DROP TABLE IF EXISTS public.timesheet_config;

ALTER TABLE IF EXISTS public.students
  ADD COLUMN IF NOT EXISTS grade_id TEXT,
  DROP CONSTRAINT IF EXISTS students__grade_id__fk,
  ADD CONSTRAINT students__grade_id__fk FOREIGN KEY (grade_id) REFERENCES public.grade(grade_id);

CREATE TABLE IF NOT EXISTS public.grade_organization (
    grade_id TEXT NOT NULL,
    grade_value INTEGER NOT NULL,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT grade_organization__pk PRIMARY KEY (grade_id, grade_value, resource_path),
    CONSTRAINT grade_organization__grade_id__fk FOREIGN KEY (grade_id) REFERENCES public.grade(grade_id)
);

CREATE POLICY rls_grade_organization ON "grade_organization"
USING (permission_check(resource_path, 'grade_organization'))
WITH CHECK (permission_check(resource_path, 'grade_organization'));

ALTER TABLE "grade_organization" ENABLE ROW LEVEL security;
ALTER TABLE "grade_organization" FORCE ROW LEVEL security;
