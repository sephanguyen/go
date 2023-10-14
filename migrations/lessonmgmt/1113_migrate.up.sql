CREATE TABLE IF NOT EXISTS public.student_packages (
  student_package_id TEXT NOT NULL,
  student_id         TEXT NOT NULL,
  package_id         TEXT NULL,
  start_at           TIMESTAMPTZ NOT NULL,
  end_at             TIMESTAMPTZ NOT NULL,
  properties         JSONB NOT NULL,
  is_active          BOOL NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL,
  updated_at         TIMESTAMPTZ NOT NULL,
  deleted_at         TIMESTAMPTZ NULL,
  resource_path      TEXT NOT NULL DEFAULT AUTOFILLRESOURCEPATH(),
  location_ids       _TEXT NULL,
  CONSTRAINT pk__student_packages PRIMARY KEY (student_package_id)
);

CREATE POLICY rls_student_packages ON "student_packages"
USING (permission_check(resource_path, 'student_packages'))
WITH CHECK (permission_check(resource_path, 'student_packages'));

CREATE POLICY rls_student_packages_restrictive ON "student_packages"
AS RESTRICTIVE TO public
USING (permission_check(resource_path, 'student_packages'))
WITH CHECK (permission_check(resource_path, 'student_packages'));

ALTER TABLE "student_packages" ENABLE ROW LEVEL security;
ALTER TABLE "student_packages" FORCE ROW LEVEL security;