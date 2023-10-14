CREATE TABLE IF NOT EXISTS public.product_group (
    product_group_id TEXT NOT NULL,
    group_name text NOT NULL,
    group_tag text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    
    CONSTRAINT pk__product_group PRIMARY KEY (product_group_id)
);

CREATE POLICY rls_product_group ON "product_group"
    using (permission_check(resource_path, 'product_group'))
    with check (permission_check(resource_path, 'product_group'));

CREATE POLICY rls_product_group_restrictive ON "product_group"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'product_group'))
    WITH CHECK (permission_check(resource_path, 'product_group'));

ALTER TABLE "product_group" ENABLE ROW LEVEL security;
ALTER TABLE "product_group" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.product_group_mapping (
    product_group_id TEXT NOT NULL,
    product_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT pk__product_group_mapping PRIMARY KEY (product_group_id, product_id),
    CONSTRAINT fk__product_group_mapping__product_group_id FOREIGN KEY(product_group_id) REFERENCES product_group(product_group_id),
    CONSTRAINT fk__product_group_mapping__product_id FOREIGN KEY(product_id) REFERENCES product(product_id)
);

CREATE POLICY rls_product_group_mapping ON "product_group_mapping"
    using (permission_check(resource_path, 'product_group_mapping'))
    with check (permission_check(resource_path, 'product_group_mapping'));

CREATE POLICY rls_product_group_mapping_restrictive ON "product_group_mapping"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'product_group_mapping'))
    WITH CHECK (permission_check(resource_path, 'product_group_mapping'));

ALTER TABLE "product_group_mapping" ENABLE ROW LEVEL security;
ALTER TABLE "product_group_mapping" FORCE ROW LEVEL security;
