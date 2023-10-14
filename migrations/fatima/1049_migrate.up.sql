CREATE TABLE IF NOT EXISTS public.package_quantity_type_mapping (
    package_type text NOT NULL,
    quantity_type text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT package_quantity_type_mapping_pk PRIMARY KEY (package_type)
);

CREATE POLICY rls_package_quantity_type_mapping ON "package_quantity_type_mapping" using (permission_check(resource_path, 'package_quantity_type_mapping')) with check (permission_check(resource_path, 'package_quantity_type_mapping'));

ALTER TABLE "package_quantity_type_mapping" ENABLE ROW LEVEL security;
ALTER TABLE "package_quantity_type_mapping" FORCE ROW LEVEL security;
