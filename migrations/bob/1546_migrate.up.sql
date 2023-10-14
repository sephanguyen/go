CREATE TYPE product_type AS ENUM ('ONE_TIME', 'SCHEDULED', 'SLOT_BASED', 'FREQUENCY');

CREATE TYPE schedule_type AS ENUM ('SCHEDULE_FIXED', 'SCHEDULE_WEEKLY');

CREATE TYPE defined_type AS ENUM ('DEFINED_BY_COURSE', 'DEFINED_BY_ORDER');

CREATE TABLE IF NOT EXISTS public.product_type_schedule (
    product_type_schedule_id TEXT NOT NULL,
    product_type product_type NOT NULL,
    schedule_type schedule_type NOT NULL,
    defined_by defined_type NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT product_type_schedule_pk PRIMARY KEY (product_type_schedule_id)
);

CREATE TABLE IF NOT EXISTS public.course_location_schedule (
    course_location_schedule_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    academic_week TEXT,
    product_type_schedule product_type,
    frequency smallint,
    total_no_lessons smallint,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT course_location_schedule_pk PRIMARY KEY (course_location_schedule_id),
    CONSTRAINT course_location_schedule_fk FOREIGN KEY (course_id, location_id) REFERENCES public.course_access_paths(course_id, location_id)
);

CREATE POLICY rls_product_type_schedule ON "product_type_schedule"
    USING (permission_check(resource_path, 'product_type_schedule'))
    WITH CHECK (permission_check(resource_path, 'product_type_schedule'));

CREATE POLICY rls_product_type_schedule_restrictive ON "product_type_schedule"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'product_type_schedule'))
    WITH CHECK (permission_check(resource_path, 'product_type_schedule'));

ALTER TABLE "product_type_schedule" ENABLE ROW LEVEL security;
ALTER TABLE "product_type_schedule" FORCE ROW LEVEL security;

CREATE POLICY rls_course_location_schedule ON "course_location_schedule"
    USING (permission_check(resource_path, 'course_location_schedule'))
    WITH CHECK (permission_check(resource_path, 'course_location_schedule'));

CREATE POLICY rls_course_location_schedule_restrictive ON "course_location_schedule"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'course_location_schedule'))
    WITH CHECK (permission_check(resource_path, 'course_location_schedule'));

ALTER TABLE "course_location_schedule" ENABLE ROW LEVEL security;
ALTER TABLE "course_location_schedule" FORCE ROW LEVEL security;
