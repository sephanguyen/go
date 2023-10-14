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
    CONSTRAINT fk__academic_week__academic_year_id FOREIGN KEY (academic_year_id) REFERENCES public.academic_year(academic_year_id),
    CONSTRAINT academic_week_name_academic_year_location_id_unique UNIQUE(name, academic_year_id, location_id)
);

-- Add CONSTRAINT unique date, location_id and academic_year_id for table academic closed day in another PR
CREATE TABLE IF NOT EXISTS public.academic_closed_day (
    academic_closed_day_id TEXT NOT NULL,
    "date" date NOT NULL,
    academic_year_id TEXT NOT NULL,
    academic_week_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT pk__academic_closed_day PRIMARY KEY (academic_closed_day_id),
    CONSTRAINT fk__academic_closed_day__academic_year_id FOREIGN KEY (academic_year_id) REFERENCES public.academic_year(academic_year_id)
);

CREATE POLICY rls_academic_year ON "academic_year"
USING (permission_check(resource_path, 'academic_year')) WITH CHECK (permission_check(resource_path, 'academic_year'));
CREATE POLICY rls_academic_year_restrictive ON "academic_year" AS RESTRICTIVE
USING (permission_check(resource_path, 'academic_year'))WITH CHECK (permission_check(resource_path, 'academic_year'));

ALTER TABLE "academic_year" ENABLE ROW LEVEL security;
ALTER TABLE "academic_year" FORCE ROW LEVEL security;

CREATE POLICY rls_academic_week ON "academic_week"
USING (permission_check(resource_path, 'academic_week')) WITH CHECK (permission_check(resource_path, 'academic_week'));
CREATE POLICY rls_academic_week_restrictive ON "academic_week" AS RESTRICTIVE
USING (permission_check(resource_path, 'academic_week'))WITH CHECK (permission_check(resource_path, 'academic_week'));

ALTER TABLE "academic_week" ENABLE ROW LEVEL security;
ALTER TABLE "academic_week" FORCE ROW LEVEL security;

CREATE POLICY rls_academic_closed_day ON "academic_closed_day"
USING (permission_check(resource_path, 'academic_closed_day')) WITH CHECK (permission_check(resource_path, 'academic_closed_day'));
CREATE POLICY rls_academic_closed_day_restrictive ON "academic_closed_day" AS RESTRICTIVE
USING (permission_check(resource_path, 'academic_closed_day'))WITH CHECK (permission_check(resource_path, 'academic_closed_day'));

ALTER TABLE "academic_closed_day" ENABLE ROW LEVEL security;
ALTER TABLE "academic_closed_day" FORCE ROW LEVEL security;
