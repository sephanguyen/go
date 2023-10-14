CREATE TABLE public.grade (
  grade_id TEXT NOT NULL,
  name text NOT NULL,
  is_archived boolean DEFAULT false NOT NULL,
  partner_internal_id VARCHAR(50) NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  created_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone NULL,
  resource_path text NOT NULL DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.grade
ADD CONSTRAINT grade_pk PRIMARY KEY (grade_id);

CREATE POLICY rls_grade ON "grade" using (permission_check(resource_path, 'grade')) with check (permission_check(resource_path, 'grade'));
ALTER TABLE "grade" ENABLE ROW LEVEL security;
ALTER TABLE "grade" FORCE ROW LEVEL security;