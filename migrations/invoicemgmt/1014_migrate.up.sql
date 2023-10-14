DROP TABLE IF EXISTS public.invoice_billing_item;
DROP POLICY IF EXISTS rls_invoice_billing_item ON "invoice_billing_item";
DROP SEQUENCE IF EXISTS public.invoice_billing_item_id_seq;

CREATE TABLE IF NOT EXISTS public.invoice_bill_item (
    invoice_bill_item_id text NOT NULL,
    invoice_id integer NOT NULL,
    bill_item_sequence_number integer NOT NULL,
    past_billing_status text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT invoice_bill_item_pk PRIMARY KEY (invoice_bill_item_id),
    CONSTRAINT invoice_bill_item_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice"(invoice_id)
);

CREATE SEQUENCE public.invoice_bill_item_id_seq
    AS integer;

ALTER SEQUENCE public.invoice_bill_item_id_seq OWNED BY public.invoice_bill_item.invoice_bill_item_id;

ALTER TABLE ONLY public.invoice_bill_item ALTER COLUMN invoice_bill_item_id SET DEFAULT nextval('public.invoice_bill_item_id_seq'::regclass);

CREATE POLICY rls_invoice_bill_item ON "invoice_bill_item" using (permission_check(resource_path, 'invoice_bill_item')) with check (permission_check(resource_path, 'invoice_bill_item'));

ALTER TABLE "invoice_bill_item" ENABLE ROW LEVEL security;
ALTER TABLE "invoice_bill_item" FORCE ROW LEVEL security;
