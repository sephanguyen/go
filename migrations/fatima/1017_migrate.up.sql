CREATE TABLE public.product_fee (
    id integer NOT NULL,
    fee_type text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT product_fee_pk PRIMARY KEY (id),
    CONSTRAINT fk__product_fee__id FOREIGN KEY (id) REFERENCES public.product (id)
);

CREATE POLICY rls_product_fee ON "product_fee" USING (permission_check(resource_path, 'product_fee')) WITH CHECK (permission_check(resource_path, 'product_fee'));

ALTER TABLE "product_fee" ENABLE ROW LEVEL security;
ALTER TABLE "product_fee" FORCE ROW LEVEL security;
