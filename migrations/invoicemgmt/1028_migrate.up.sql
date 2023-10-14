CREATE TABLE public.invoice_schedule (
    invoice_schedule_id text NOT NULL,
    invoice_date timestamp with time zone NOT NULL,
    status text NOT NULL,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT null,
    CONSTRAINT invoice_schedule_pk PRIMARY KEY (invoice_schedule_id)
);

CREATE POLICY rls_invoice_schedule ON "invoice_schedule" USING (permission_check(resource_path, 'invoice_schedule')) WITH CHECK (permission_check(resource_path, 'invoice_schedule'));

ALTER TABLE "invoice_schedule" ENABLE ROW LEVEL security;
ALTER TABLE "invoice_schedule" FORCE ROW LEVEL security;