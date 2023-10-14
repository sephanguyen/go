CREATE TABLE IF NOT EXISTS public.payment (
    payment_id integer NOT NULL,
    invoice_id integer NOT NULL,
    payment_method text NOT NULL,
    payment_due_date timestamp with time zone NOT NULL,
    payment_expiry_date timestamp with time zone NOT NULL,
    payment_date timestamp with time zone NOT NULL,
    payment_status text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT payment_pk PRIMARY KEY (payment_id),
    CONSTRAINT payment_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice"(invoice_id)
);

CREATE POLICY rls_payment ON "payment" USING (permission_check(resource_path, 'payment')) WITH CHECK (permission_check(resource_path, 'payment'));

ALTER TABLE "payment" ENABLE ROW LEVEL security;
ALTER TABLE "payment" FORCE ROW LEVEL security;
