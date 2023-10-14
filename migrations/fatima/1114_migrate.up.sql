CREATE TABLE public.upcoming_student_package_by_order (
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

CREATE POLICY rls_upcoming_student_package_by_order ON "upcoming_student_package_by_order"
    USING (permission_check(resource_path, 'upcoming_student_package_by_order'))
    WITH CHECK (permission_check(resource_path, 'upcoming_student_package_by_order'));

CREATE POLICY rls_upcoming_student_package_by_order_restrictive ON "upcoming_student_package_by_order"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'upcoming_student_package_by_order'))
    WITH CHECK (permission_check(resource_path, 'upcoming_student_package_by_order'));

ALTER TABLE "upcoming_student_package_by_order" ENABLE ROW LEVEL security;
ALTER TABLE "upcoming_student_package_by_order" FORCE ROW LEVEL security;

CREATE TABLE public.upcoming_student_course (
    upcoming_student_package_id text NOT NULL,
    student_id text NOT NULL,
    course_id text NULL,
    location_id text NOT NULL,
    student_package_id text NOT NULL,
    student_start_date timestamp with time zone NOT NULL,
    student_end_date timestamp with time zone NOT NULL,
    course_slot integer,
    course_slot_per_week integer,
    weight integer,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    package_type text,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT fk_upcoming_student_course_upcoming_student_package_by_order FOREIGN KEY (upcoming_student_package_id) REFERENCES public.upcoming_student_package_by_order(upcoming_student_package_id),
    CONSTRAINT upcoming_student_course_pk PRIMARY KEY (upcoming_student_package_id,student_id,course_id,location_id,student_package_id)
);

CREATE POLICY rls_upcoming_student_course ON "upcoming_student_course"
    USING (permission_check(resource_path, 'upcoming_student_course'))
    WITH CHECK (permission_check(resource_path, 'upcoming_student_course'));

CREATE POLICY rls_upcoming_student_course_restrictive ON "upcoming_student_course"  AS RESTRICTIVE TO PUBLIC
    USING (permission_check(resource_path, 'upcoming_student_course'))
    WITH CHECK (permission_check(resource_path, 'upcoming_student_course'));

ALTER TABLE "upcoming_student_course" ENABLE ROW LEVEL security;
ALTER TABLE "upcoming_student_course" FORCE ROW LEVEL security;