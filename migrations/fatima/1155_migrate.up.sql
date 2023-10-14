CREATE TABLE public.student_discount_tracker (
    discount_tracker_id text NOT NULL,
    student_id text NOT NULL,
    location_id text NOT NULL,
    student_product_id text NOT NULL,
    product_id text NOT NULL,
    discount_type text DEFAULT NULL,
    discount_status text DEFAULT NULL,
    discount_start_date timestamp with time zone,
    discount_end_date timestamp with time zone,
    student_product_start_date timestamp with time zone,
    student_product_end_date timestamp with time zone,
    student_product_status text DEFAULT NULL,
    updated_from_discount_tracker_id text DEFAULT NULL,
    updated_to_discount_tracker_id text DEFAULT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE public.student_discount_tracker
    ADD CONSTRAINT student_discount_tracker_pk PRIMARY KEY (discount_tracker_id);

ALTER TABLE public.student_discount_tracker
    ADD CONSTRAINT fk_student_discount_tracker_student_id FOREIGN KEY (student_id) REFERENCES public.students(student_id);

ALTER TABLE public.student_discount_tracker
    ADD CONSTRAINT fk_student_discount_tracker_location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id);

ALTER TABLE public.student_discount_tracker
    ADD CONSTRAINT fk_student_discount_tracker_student_product_id FOREIGN KEY (student_product_id) REFERENCES public.student_product(student_product_id);

CREATE POLICY rls_student_discount_tracker ON "student_discount_tracker"
    using (permission_check(resource_path, 'student_discount_tracker'))
    with check (permission_check(resource_path, 'student_discount_tracker'));

CREATE POLICY rls_student_discount_tracker_restrictive ON "student_discount_tracker"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'student_discount_tracker'))
    WITH CHECK (permission_check(resource_path, 'student_discount_tracker'));

ALTER TABLE "student_discount_tracker" ENABLE ROW LEVEL security;
ALTER TABLE "student_discount_tracker" FORCE ROW LEVEL security;
