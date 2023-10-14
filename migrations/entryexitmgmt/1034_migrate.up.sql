CREATE TABLE IF NOT EXISTS public.grade
(
    grade_id TEXT NOT NULL,
    name TEXT NOT NULL,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    partner_internal_id VARCHAR(50) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),
    "sequence" INTEGER,
    remarks TEXT,

    CONSTRAINT grade_pk PRIMARY KEY (grade_id)
);

CREATE POLICY rls_grade ON "grade"
USING (permission_check(resource_path, 'grade')) WITH CHECK (permission_check(resource_path, 'grade'));

CREATE POLICY rls_grade_restrictive ON "grade" AS RESTRICTIVE
USING (permission_check(resource_path, 'grade'))WITH CHECK (permission_check(resource_path, 'grade'));

ALTER TABLE "grade" ENABLE ROW LEVEL security;
ALTER TABLE "grade" FORCE ROW LEVEL security;