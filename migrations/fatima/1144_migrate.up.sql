ALTER TABLE IF EXISTS public.upcoming_student_course DROP CONSTRAINT IF EXISTS fk_upcoming_student_course_upcoming_student_package_by_order;
DROP TABLE IF EXISTS public.upcoming_student_package_by_order;

CREATE TABLE IF NOT EXISTS public.upcoming_student_package  (
                                                          upcoming_student_package_id text NOT NULL,
                                                          student_package_id text NOT NULL,
                                                          student_id text NOT NULL,
                                                          package_id text NOT NULL,
                                                          start_at timestamp with time zone NOT NULL,
                                                          end_at timestamp with time zone NOT NULL,
                                                          properties JSONB NOT NULL,
                                                          is_active boolean NOT NULL,
                                                          updated_at timestamp with time zone NOT NULL,
                                                          created_at timestamp with time zone NOT NULL,
                                                          deleted_at timestamp with time zone,
                                                          resource_path text DEFAULT autofillresourcepath(),
                                                          location_ids text[],
                                                          student_subscription_student_package_id text DEFAULT NULL,
                                                          CONSTRAINT upcoming_student_package_id_pk PRIMARY KEY (upcoming_student_package_id)
);

CREATE POLICY rls_upcoming_student_package ON "upcoming_student_package"
    USING (permission_check(resource_path, 'upcoming_student_package'))
    WITH CHECK (permission_check(resource_path, 'upcoming_student_package'));

CREATE POLICY rls_upcoming_student_package_restrictive ON "upcoming_student_package"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'upcoming_student_package'))
    WITH CHECK (permission_check(resource_path, 'upcoming_student_package'));

ALTER TABLE "upcoming_student_package" ENABLE ROW LEVEL security;
ALTER TABLE "upcoming_student_package" FORCE ROW LEVEL security;
