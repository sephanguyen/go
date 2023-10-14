CREATE TABLE IF NOT EXISTS public.student_package_order (
    student_package_order_id text NOT NULL,
    user_id text NOT NULL,
    order_id text NOT NULL,
    course_id text NOT NULL,
    start_at timestamp with time zone NOT NULL,
    end_at timestamp with time zone NOT NULL,
    student_package_object JSONB NOT NULL,
    student_package_id text  NOT NULL,
    upcoming_student_package_id text,
    is_current_student_package boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT student_package_order_id__pk PRIMARY KEY (student_package_order_id),
    CONSTRAINT fk_student_package_order_order_id FOREIGN KEY (order_id) REFERENCES public.order(order_id),
    CONSTRAINT fk_student_package_order_student_package_id FOREIGN KEY (student_package_id) REFERENCES public.student_packages(student_package_id),
    CONSTRAINT fk_student_package_order_upcoming_student_package_id FOREIGN KEY (upcoming_student_package_id) REFERENCES public.upcoming_student_package(upcoming_student_package_id)
);

CREATE POLICY rls_student_package_order ON "student_package_order"
    USING (permission_check(resource_path, 'student_package_order'))
    WITH CHECK (permission_check(resource_path, 'student_package_order'));

CREATE POLICY rls_student_package_order_restrictive ON "student_package_order"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'student_package_order'))
    WITH CHECK (permission_check(resource_path, 'student_package_order'));

ALTER TABLE "student_package_order" ENABLE ROW LEVEL security;
ALTER TABLE "student_package_order" FORCE ROW LEVEL security;
