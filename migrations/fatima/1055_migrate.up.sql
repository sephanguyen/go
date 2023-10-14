CREATE TABLE public.student_product (
    student_product_id text NOT NULL,
    student_id text NOT NULL,
    product_id integer NOT NULL,
    upcoming_billing_date timestamp with time zone,
    amount numeric(12,2) NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    product_status text NULL,
    approval_status text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    location_id text NOT NULL,
    CONSTRAINT student_product_pk PRIMARY KEY (student_product_id),
    CONSTRAINT fk_student_product_student_id FOREIGN KEY (student_id) REFERENCES public.students (student_id),
    CONSTRAINT fk_student_product_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id),
    CONSTRAINT fk_student_product_location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id)
);

CREATE POLICY rls_student_product ON "student_product" USING (permission_check(resource_path, 'student_product')) WITH CHECK (permission_check(resource_path, 'student_product'));

ALTER TABLE "student_product" ENABLE ROW LEVEL security;
ALTER TABLE "student_product" FORCE ROW LEVEL security;

CREATE TABLE public.student_package_by_order (
    student_package_id text NOT NULL,
    student_id text NOT NULL,
    package_id integer NOT NULL,
    start_at timestamp with time zone NOT NULL,
    end_at timestamp with time zone NOT NULL,
    properties JSONB NOT NULL,
    is_active boolean NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    location_ids text[],
    CONSTRAINT student_package_by_order_pk PRIMARY KEY (student_package_id),
    CONSTRAINT fk_student_package_student_id FOREIGN KEY (student_id) REFERENCES public.students (student_id),
    CONSTRAINT fk_student_package_package_id FOREIGN KEY (package_id) REFERENCES public.package(package_id)
);

CREATE INDEX idx__student_package_by_order__student_id__start_at__end_at ON public.student_package_by_order USING btree (student_id, start_at, end_at);
CREATE INDEX IF NOT EXISTS idx__student_package_by_order__properties__can_do_quiz ON public.student_package_by_order USING gin ((properties->'can_do_quiz'));
CREATE POLICY rls_student_package_by_order ON "student_package_by_order" USING (permission_check(resource_path, 'student_package_by_order')) WITH CHECK (permission_check(resource_path, 'student_package_by_order'));

ALTER TABLE "student_package_by_order" ENABLE ROW LEVEL security;
ALTER TABLE "student_package_by_order" FORCE ROW LEVEL security;