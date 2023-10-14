CREATE TABLE public.product_course (
                                       product_id integer NOT NULL,
                                       course_id text NOT NULL,
                                       mandatory_flag boolean DEFAULT false NOT NULL,
                                       course_weight integer DEFAULT 1 NOT NULL,
                                       created_at timestamp with time zone NOT NULL,
                                       resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.product_course
    ADD CONSTRAINT product_course_pk PRIMARY KEY (product_id,course_id);

ALTER TABLE public.product_course ADD CONSTRAINT fk_product_course_id FOREIGN KEY(product_id) REFERENCES product(id);

CREATE POLICY rls_product_course ON "product_course" using (permission_check(resource_path, 'product_course')) with check (permission_check(resource_path, 'product_course'));

ALTER TABLE "product_course" ENABLE ROW LEVEL security;
ALTER TABLE "product_course" FORCE ROW LEVEL security;
