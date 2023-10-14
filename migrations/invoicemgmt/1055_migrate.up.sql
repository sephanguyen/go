CREATE TABLE IF NOT EXISTS bulk_payment_validations_detail (
    bulk_payment_validations_detail_id text NOT NULL,
    bulk_payment_validations_id text NOT NULL,
    invoice_id text NOT NULL,
    payment_id text NOT NULL,
    validated_result_code text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT bulk_payment_validations_detail__pk PRIMARY KEY (bulk_payment_validations_detail_id),
    CONSTRAINT bulk_payment_validations_detail_bulk_payment_validations_fk FOREIGN KEY (bulk_payment_validations_id) REFERENCES "bulk_payment_validations"(bulk_payment_validations_id),
    CONSTRAINT bulk_payment_validations_detail_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice"(invoice_id),
    CONSTRAINT bulk_payment_validations_detail_payment_fk FOREIGN KEY (payment_id) REFERENCES "payment"(payment_id)
);

CREATE POLICY rls_bulk_payment_validations_detail ON "bulk_payment_validations_detail"
USING (permission_check(resource_path, 'bulk_payment_validations_detail'))
WITH CHECK (permission_check(resource_path, 'bulk_payment_validations_detail'));

CREATE POLICY rls_bulk_payment_validations_detail_restrictive ON "bulk_payment_validations_detail" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'bulk_payment_validations_detail'))
WITH CHECK (permission_check(resource_path, 'bulk_payment_validations_detail'));

ALTER TABLE "bulk_payment_validations_detail" ENABLE ROW LEVEL security;
ALTER TABLE "bulk_payment_validations_detail" FORCE ROW LEVEL security;