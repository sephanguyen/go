CREATE TABLE public.product_accounting_category (
    product_id integer NOT NULL,
    accounting_category_id integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.product_accounting_category
    ADD CONSTRAINT product_accounting_category_pk PRIMARY KEY (product_id, accounting_category_id);

ALTER TABLE public.product_accounting_category ADD CONSTRAINT fk_product_id FOREIGN KEY(product_id) REFERENCES product(id);
ALTER TABLE public.product_accounting_category ADD CONSTRAINT fk_accounting_category_id FOREIGN KEY(accounting_category_id) REFERENCES accounting_category(id);

CREATE POLICY rls_product_accounting_category ON "product_accounting_category" using (permission_check(resource_path, 'product_accounting_category')) with check (permission_check(resource_path, 'product_accounting_category'));

ALTER TABLE "product_accounting_category" ENABLE ROW LEVEL security;
ALTER TABLE "product_accounting_category" FORCE ROW LEVEL security;