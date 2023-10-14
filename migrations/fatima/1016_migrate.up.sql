CREATE TABLE public.product_material (
    id integer NOT NULL,
    material_type text NOT NULL,
    custom_billing_date timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT product_material_pk PRIMARY KEY (id),
    CONSTRAINT fk__product_material__id FOREIGN KEY (id) REFERENCES public.product (id)
);

CREATE POLICY rls_product_material ON "product_material" USING (permission_check(resource_path, 'product_material')) WITH CHECK (permission_check(resource_path, 'product_material'));

ALTER TABLE "product_material" ENABLE ROW LEVEL security;
ALTER TABLE "product_material" FORCE ROW LEVEL security;
