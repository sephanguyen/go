CREATE TABLE IF NOT EXISTS public.bulk_payment_validations (
    bulk_payment_validations_id text NOT NULL,
    payment_method text NOT NULL,
    successful_validations int NOT NULL,
    failed_validations int NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,

    CONSTRAINT bulk_payment_validations_pk PRIMARY KEY (bulk_payment_validations_id)
);

CREATE POLICY rls_bulk_payment_validations ON "bulk_payment_validations" USING (permission_check(resource_path, 'bulk_payment_validations')) with check (permission_check(resource_path, 'bulk_payment_validations'));

CREATE POLICY rls_bulk_payment_validations_restrictive ON "bulk_payment_validations" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'bulk_payment_validations')) WITH CHECK (permission_check(resource_path, 'bulk_payment_validations'));

ALTER TABLE "bulk_payment_validations" ENABLE ROW LEVEL security;
ALTER TABLE "bulk_payment_validations" FORCE ROW LEVEL security;