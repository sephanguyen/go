CREATE TABLE IF NOT EXISTS public.academic_week (
    academic_week_id TEXT NOT NULL,
    name TEXT NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    period TEXT NOT NULL,
    academic_year_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT pk__academic_week PRIMARY KEY (academic_week_id),
    CONSTRAINT academic_week_name_academic_year_location_id_unique UNIQUE(name, academic_year_id, location_id)
);

CREATE POLICY rls_academic_week ON "academic_week"
USING (permission_check(resource_path, 'academic_week')) WITH CHECK (permission_check(resource_path, 'academic_week'));
CREATE POLICY rls_academic_week_restrictive ON "academic_week" AS RESTRICTIVE
USING (permission_check(resource_path, 'academic_week'))WITH CHECK (permission_check(resource_path, 'academic_week'));

ALTER TABLE "academic_week" ENABLE ROW LEVEL security;
ALTER TABLE "academic_week" FORCE ROW LEVEL security;
