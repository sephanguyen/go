CREATE TABLE public.product_package (
    id integer NOT NULL,
    package_type text NOT NULL,
    max_slot integer NOT NULL,
    package_start_date timestamp with time zone,
    package_end_date timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.product_package
    ADD CONSTRAINT product_package_pk PRIMARY KEY (id);

ALTER TABLE public.product_package ADD CONSTRAINT fk_product_package_id FOREIGN KEY(id) REFERENCES product(id);
ALTER TABLE public.product ADD CONSTRAINT fk_product_tax_id FOREIGN KEY(tax_id) REFERENCES tax(id);
ALTER TABLE public.product ADD CONSTRAINT fk_product_billing_schedule_id FOREIGN KEY(billing_schedule_id) REFERENCES billing_schedule(id);

CREATE POLICY rls_product_package ON "product_package" USING (permission_check(resource_path, 'product_package')) WITH CHECK (permission_check(resource_path, 'package'));

ALTER TABLE "product_package" ENABLE ROW LEVEL security;
ALTER TABLE "product_package" FORCE ROW LEVEL security;

ALTER TABLE "product" RENAME available_util TO available_until;
