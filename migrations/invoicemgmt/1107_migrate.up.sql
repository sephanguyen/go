CREATE TABLE IF NOT EXISTS public.bulk_payment (
    bulk_payment_id TEXT NOT NULL,
    bulk_payment_status TEXT NOT NULL,
    payment_method TEXT NOT NULL,
    invoice_status TEXT NOT NULL,
    invoice_type TEXT[] NOT NULL,
    payment_status TEXT[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__bulk_payment PRIMARY KEY (bulk_payment_id)
);

CREATE POLICY rls_bulk_payment ON "bulk_payment"
USING (permission_check(resource_path, 'bulk_payment')) WITH CHECK (permission_check(resource_path, 'bulk_payment'));

CREATE POLICY rls_bulk_payment_restrictive ON "bulk_payment" AS RESTRICTIVE
USING (permission_check(resource_path, 'bulk_payment'))WITH CHECK (permission_check(resource_path, 'bulk_payment'));

ALTER TABLE "bulk_payment" ENABLE ROW LEVEL security;
ALTER TABLE "bulk_payment" FORCE ROW LEVEL security;
