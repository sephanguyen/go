CREATE TABLE IF NOT EXISTS public.invoice_billing_item (
    invoice_billing_item_id integer NOT NULL,
    invoice_id integer NOT NULL,
    bill_item_sequence_number integer NOT NULL,
    past_billing_status text NOT NULL, -- pending, billed
    resource_path text DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT invoice_billing_item_pk PRIMARY KEY (invoice_billing_item_id),
    CONSTRAINT invoice_billing_item_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice"(invoice_id)
);

CREATE SEQUENCE public.invoice_billing_item_id_seq
    AS integer;

ALTER SEQUENCE public.invoice_billing_item_id_seq OWNED BY public.invoice_billing_item.invoice_billing_item_id;

ALTER TABLE ONLY public.invoice_billing_item ALTER COLUMN invoice_billing_item_id SET DEFAULT nextval('public.invoice_billing_item_id_seq'::regclass);

CREATE POLICY rls_invoice_billing_item ON "invoice_billing_item" using (permission_check(resource_path, 'invoice_billing_item')) with check (permission_check(resource_path, 'invoice_billing_item'));

ALTER TABLE "invoice_billing_item" ENABLE ROW LEVEL security;
ALTER TABLE "invoice_billing_item" FORCE ROW LEVEL security;