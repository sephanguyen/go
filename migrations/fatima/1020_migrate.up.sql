CREATE TABLE public.product_grade (
    product_id integer NOT NULL,
    grade_id integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.product_grade
    ADD CONSTRAINT product_grade_pk PRIMARY KEY (product_id, grade_id);

ALTER TABLE public.product_grade ADD CONSTRAINT fk_product_id FOREIGN KEY(product_id) REFERENCES public.product(id);
ALTER TABLE public.product_grade ADD CONSTRAINT fk_grade_id FOREIGN KEY(grade_id) REFERENCES public.grade(id);

CREATE POLICY rls_product_grade ON "product_grade" using (permission_check(resource_path, 'product_grade')) with check (permission_check(resource_path, 'product_grade'));

ALTER TABLE "product_grade" ENABLE ROW LEVEL security;
ALTER TABLE "product_grade" FORCE ROW LEVEL security;
