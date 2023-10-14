CREATE TABLE public.product_location (
    product_id integer NOT NULL,
    location_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.product_location ADD CONSTRAINT product_location_pk PRIMARY KEY (product_id, location_id);

ALTER TABLE public.product_location ADD CONSTRAINT fk_product_location_id FOREIGN KEY(product_id) REFERENCES product(id);

CREATE POLICY rls_product_location ON "product_location" using (permission_check(resource_path, 'product_location')) with check (permission_check(resource_path, 'product_location'));

ALTER TABLE "product_location" ENABLE ROW LEVEL security;
ALTER TABLE "product_location" FORCE ROW LEVEL security;
