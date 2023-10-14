-- =========================================================================================
-- ============================== public.package_course table ==============================
-- =========================================================================================
CREATE TABLE IF NOT EXISTS public.package_course (
    package_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    mandatory_flag BOOLEAN NOT NULL DEFAULT false,
    course_weight int4 NOT NULL DEFAULT 1,
    max_slots_per_course TEXT NOT NULL DEFAULT 1,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT fk__package_course__course_id FOREIGN KEY (course_id) REFERENCES public.courses(course_id),
    CONSTRAINT pk__package_course PRIMARY KEY (package_id,course_id)
);

CREATE POLICY rls_package_course ON "package_course"
USING (permission_check(resource_path, 'package_course')) WITH CHECK (permission_check(resource_path, 'package_course'));
CREATE POLICY rls_package_course_restrictive ON "package_course" AS RESTRICTIVE
USING (permission_check(resource_path, 'package_course'))WITH CHECK (permission_check(resource_path, 'package_course'));

ALTER TABLE "package_course" ENABLE ROW LEVEL security;
ALTER TABLE "package_course" FORCE ROW LEVEL security;

-- =========================================================================================
-- ============================== public.product table =====================================
-- =========================================================================================
CREATE TABLE IF NOT EXISTS public.product (
    product_id TEXT NOT NULL,
    "name" TEXT NOT NULL,
    product_type TEXT NOT NULL,
    tax_id TEXT,
    available_from timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    available_until timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    remarks TEXT,
    custom_billing_period TIMESTAMP WITH TIME ZONE,
    billing_schedule_id TEXT,
    disable_pro_rating_flag BOOLEAN NOT NULL DEFAULT false,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    is_unique BOOLEAN DEFAULT false,
    
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT pk__product PRIMARY KEY (product_id)
);

CREATE POLICY rls_product ON "product" USING (permission_check(resource_path, 'product')) WITH CHECK (permission_check(resource_path, 'product'));
CREATE POLICY rls_product_restrictive ON "product" AS RESTRICTIVE USING (permission_check(resource_path, 'product')) WITH CHECK (permission_check(resource_path, 'product'));

ALTER TABLE "product" ENABLE ROW LEVEL security;
ALTER TABLE "product" FORCE ROW LEVEL security;

-- =========================================================================================
-- ============================== public.student_product table =============================
-- =========================================================================================
CREATE TABLE IF NOT EXISTS public.student_product (
    student_product_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    upcoming_billing_date TIMESTAMP WITH TIME ZONE ,
    "start_date" TIMESTAMP WITH TIME ZONE ,
    end_date TIMESTAMP WITH TIME ZONE ,
    product_status TEXT NOT NULL,
    approval_status TEXT ,
    updated_from_student_product_id TEXT ,
    updated_to_student_product_id TEXT ,
    student_product_label TEXT ,
    is_unique BOOLEAN DEFAULT false ,
    root_student_product_id TEXT ,
    
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT fk__student_product__location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id),
    CONSTRAINT fk__student_product__product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id),
    CONSTRAINT fk__student_product__student_id FOREIGN KEY (student_id) REFERENCES public.students(student_id),
    CONSTRAINT fk__student_product__updated_from_student_product_id FOREIGN KEY (updated_from_student_product_id) REFERENCES public.student_product(student_product_id),
    CONSTRAINT fk__student_product__updated_to_student_product_id FOREIGN KEY (updated_to_student_product_id) REFERENCES public.student_product(student_product_id),
    CONSTRAINT pk__student_product PRIMARY KEY (student_product_id)
);

CREATE POLICY rls_student_product ON "student_product" USING (permission_check(resource_path, 'student_product')) WITH CHECK (permission_check(resource_path, 'student_product'));
CREATE POLICY rls_student_product_restrictive ON "student_product" AS RESTRICTIVE USING (permission_check(resource_path, 'student_product')) WITH CHECK (permission_check(resource_path, 'student_product'));

ALTER TABLE "student_product" ENABLE ROW LEVEL security;
ALTER TABLE "student_product" FORCE ROW LEVEL security;
