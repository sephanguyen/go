CREATE TABLE public.product_discount (
    discount_id text NOT NULL,
    product_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.product_discount
    ADD CONSTRAINT product_discount_pk PRIMARY KEY (discount_id, product_id);

ALTER TABLE public.product_discount ADD CONSTRAINT fk_discount_id FOREIGN KEY(discount_id) REFERENCES public.discount(discount_id);
ALTER TABLE public.product_discount ADD CONSTRAINT fk_product_id FOREIGN KEY(product_id) REFERENCES public.product(product_id);

CREATE POLICY rls_product_discount ON "product_discount" USING (permission_check(resource_path, 'product_discount')) WITH CHECK (permission_check(resource_path, 'product_discount'));

ALTER TABLE "product_discount" ENABLE ROW LEVEL security;
ALTER TABLE "product_discount" FORCE ROW LEVEL security;