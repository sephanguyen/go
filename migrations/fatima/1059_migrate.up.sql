CREATE TABLE IF NOT EXISTS public.product_setting (
    product_id int NOT NULL,
    is_required_for_enrollment boolean DEFAULT false,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT product_settings_pk PRIMARY KEY (product_id),
    CONSTRAINT fk_product_setting_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id)
);

CREATE POLICY rls_product_setting ON "product_setting" using (permission_check(resource_path, 'product_setting')) with check (permission_check(resource_path, 'product_setting'));

ALTER TABLE "product_setting" ENABLE ROW LEVEL security;
ALTER TABLE "product_setting" FORCE ROW LEVEL security;
