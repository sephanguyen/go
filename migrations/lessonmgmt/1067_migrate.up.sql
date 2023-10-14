CREATE TABLE IF NOT EXISTS public.academic_year (
    academic_year_id TEXT NOT NULL,
    name TEXT NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT pk__academic_year PRIMARY KEY (academic_year_id)
);

CREATE POLICY rls_academic_year ON "academic_year"
USING (permission_check(resource_path, 'academic_year')) WITH CHECK (permission_check(resource_path, 'academic_year'));
CREATE POLICY rls_academic_year_restrictive ON "academic_year" AS RESTRICTIVE
USING (permission_check(resource_path, 'academic_year'))WITH CHECK (permission_check(resource_path, 'academic_year'));

ALTER TABLE "academic_year" ENABLE ROW LEVEL security;
ALTER TABLE "academic_year" FORCE ROW LEVEL security;
