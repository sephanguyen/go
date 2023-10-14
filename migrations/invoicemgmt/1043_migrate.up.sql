CREATE TABLE IF NOT EXISTS public.bulk_payment_request_file_payment (
    bulk_payment_request_file_payment_id text NOT NULL,
    bulk_payment_request_file_id text NOT NULL,
    payment_id text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT bulk_payment_request_file_payment__pk PRIMARY KEY (bulk_payment_request_file_payment_id),
    CONSTRAINT bulk_payment_request_file_payment_bulk_payment_request_file__fk FOREIGN KEY (bulk_payment_request_file_id) REFERENCES "bulk_payment_request_file"(bulk_payment_request_file_id),
    CONSTRAINT bulk_payment_request_file_payment_payment__fk FOREIGN KEY (payment_id) REFERENCES "payment"(payment_id)
);

CREATE POLICY rls_bulk_payment_request_file_payment ON "bulk_payment_request_file_payment"
USING (permission_check(resource_path, 'bulk_payment_request_file_payment'))
WITH CHECK (permission_check(resource_path, 'bulk_payment_request_file_payment'));

CREATE POLICY rls_bulk_payment_request_file_payment_restrictive ON "bulk_payment_request_file_payment" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'bulk_payment_request_file_payment'))
WITH CHECK (permission_check(resource_path, 'bulk_payment_request_file_payment'));

ALTER TABLE "bulk_payment_request_file_payment" ENABLE ROW LEVEL security;
ALTER TABLE "bulk_payment_request_file_payment" FORCE ROW LEVEL security;