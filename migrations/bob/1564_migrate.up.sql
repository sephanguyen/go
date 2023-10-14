CREATE TABLE IF NOT EXISTS public.reserve_class (
    reserve_class_id TEXT NOT NULL,
    student_package_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    class_id TEXT NOT NULL,
    effective_date date NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT reserve_class_pk PRIMARY KEY (reserve_class_id)
);

CREATE POLICY rls_reserve_class ON "reserve_class"
USING (permission_check(resource_path, 'reserve_class')) WITH CHECK (permission_check(resource_path, 'reserve_class'));
CREATE POLICY rls_reserve_class_restrictive ON "reserve_class" AS RESTRICTIVE
USING (permission_check(resource_path, 'reserve_class'))WITH CHECK (permission_check(resource_path, 'reserve_class'));

ALTER TABLE "reserve_class" ENABLE ROW LEVEL security;
ALTER TABLE "reserve_class" FORCE ROW LEVEL security;
