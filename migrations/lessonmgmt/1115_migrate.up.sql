CREATE TABLE IF NOT EXISTS public.course_type (
  course_type_id TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  resource_path TEXT DEFAULT autofillresourcepath(),
  CONSTRAINT course_type__pk PRIMARY KEY (course_type_id)
);

CREATE POLICY rls_course_type ON "course_type"
USING (permission_check(resource_path, 'course_type'))
WITH CHECK (permission_check(resource_path, 'course_type'));

CREATE POLICY rls_course_type_restrictive ON "course_type" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'course_type'))
with check (permission_check(resource_path, 'course_type'));


ALTER TABLE IF EXISTS public.course_type ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.course_type FORCE ROW LEVEL security;
