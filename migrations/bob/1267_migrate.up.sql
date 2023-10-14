DROP TABLE IF EXISTS public.grade_organization;

CREATE TABLE IF NOT EXISTS public.grade_organization (
    grade_organization_id TEXT NOT NULL,
    grade_id TEXT NOT NULL,
    grade_value INTEGER NOT NULL,

    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT grade_organization__pk PRIMARY KEY (grade_organization_id),
    CONSTRAINT grade_organization__grade_id__fk FOREIGN KEY (grade_id) REFERENCES public.grade(grade_id)
);

CREATE POLICY rls_grade_organization ON "grade_organization"
USING (permission_check(resource_path, 'grade_organization'))
WITH CHECK (permission_check(resource_path, 'grade_organization'));

ALTER TABLE "grade_organization" ENABLE ROW LEVEL security;
ALTER TABLE "grade_organization" FORCE ROW LEVEL security;
