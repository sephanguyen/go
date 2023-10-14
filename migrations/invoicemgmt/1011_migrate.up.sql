CREATE TABLE public.student_product (
    student_product_id text NOT NULL,
    student_id text NOT NULL,
    product_id integer NOT NULL,
    upcoming_billing_date timestamp with time zone,
    amount numeric(12,2) NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    product_status text NOT NULL,
    approval_status text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    location_id text NOT NULL,
    CONSTRAINT student_product_pk PRIMARY KEY (student_product_id),
    CONSTRAINT fk_student_product_student_id FOREIGN KEY (student_id) REFERENCES public.students (student_id),
    CONSTRAINT fk_student_product_location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id)
);

CREATE POLICY rls_student_product ON "student_product" USING (permission_check(resource_path, 'student_product')) WITH CHECK (permission_check(resource_path, 'student_product'));

ALTER TABLE "student_product" ENABLE ROW LEVEL security;
ALTER TABLE "student_product" FORCE ROW LEVEL security;

ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_student_product_id FOREIGN KEY(student_product_id) REFERENCES public.student_product(student_product_id);