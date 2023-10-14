CREATE TABLE IF NOT EXISTS public.billing_address (
    billing_address_id text NOT NULL,
    user_id text NOT NULL,
    student_payment_detail_id text NOT NULL,
    postal_code text NOT NULL,
    prefecture_name text NOT NULL,
    city text NOT NULL,
    street1 text NOT NULL,
    street2 text,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT billing_address__pk PRIMARY KEY (billing_address_id),
    CONSTRAINT billing_address_users_fk FOREIGN KEY (user_id) REFERENCES "users"(user_id),
    CONSTRAINT billing_address_student_payment_detail_fk FOREIGN KEY (student_payment_detail_id) REFERENCES "student_payment_detail"(student_payment_detail_id)
);

CREATE POLICY rls_billing_address ON "billing_address"
USING (permission_check(resource_path, 'billing_address'))
WITH CHECK (permission_check(resource_path, 'billing_address'));

CREATE POLICY rls_billing_address_restrictive ON "billing_address" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'billing_address'))
WITH CHECK (permission_check(resource_path, 'billing_address'));

ALTER TABLE "billing_address" ENABLE ROW LEVEL security;
ALTER TABLE "billing_address" FORCE ROW LEVEL security;