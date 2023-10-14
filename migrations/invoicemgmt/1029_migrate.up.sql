CREATE TABLE public.invoice_schedule_history (
    invoice_schedule_history_id text NOT NULL,
    invoice_schedule_id text NOT NULL,
    number_of_failed_invoices integer NOT NULL,
    number_of_students_without_bill_items integer NOT NULL,
    total_students integer NOT NULL,
    execution_start_date timestamp with time zone NOT NULL,
    execution_end_date timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT invoice_schedule_history_pk PRIMARY KEY (invoice_schedule_history_id),
    CONSTRAINT invoice_schedule_history_invoice_schedule_fk FOREIGN KEY (invoice_schedule_id) REFERENCES "invoice_schedule"(invoice_schedule_id)
);

CREATE POLICY rls_invoice_schedule_history ON "invoice_schedule_history" USING (permission_check(resource_path, 'invoice_schedule_history')) WITH CHECK (permission_check(resource_path, 'invoice_schedule_history'));

ALTER TABLE "invoice_schedule_history" ENABLE ROW LEVEL security;
ALTER TABLE "invoice_schedule_history" FORCE ROW LEVEL security;

ALTER TABLE public.invoice_schedule_history ADD CONSTRAINT invoice_schedule_history_invoice_schedule_id_key UNIQUE(invoice_schedule_id);