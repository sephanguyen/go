CREATE TABLE IF NOT EXISTS public.invoice_credit_note (
    credit_note_id text NOT NULL,
    invoice_id text NOT NULL,
    credit_note_sequence_number int NOT NULL,
    reason text NOT NULL,
    price numeric(12,2) NOT NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT invoice_credit_note_pk PRIMARY KEY (credit_note_id),
    CONSTRAINT invoice_credit_note_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice"(invoice_id)
);

CREATE POLICY rls_invoice_credit_note ON "invoice_credit_note" USING (permission_check(resource_path, 'invoice_credit_note')) WITH CHECK (permission_check(resource_path, 'invoice_credit_note'));

ALTER TABLE "invoice_credit_note" ENABLE ROW LEVEL security;
ALTER TABLE "invoice_credit_note" FORCE ROW LEVEL security; 

CREATE SEQUENCE public.credit_note_sequence_number_seq 
    AS integer;
ALTER TABLE ONLY public.invoice_credit_note
    ALTER COLUMN credit_note_sequence_number SET DEFAULT nextval('public.credit_note_sequence_number_seq'::regclass);
ALTER TABLE ONLY public.invoice_credit_note
    ADD CONSTRAINT credit_note_sequence_number_resource_path_unique UNIQUE (credit_note_sequence_number,resource_path);
ALTER SEQUENCE public.credit_note_sequence_number_seq OWNED 
    BY public.invoice_credit_note.credit_note_sequence_number;

CREATE OR REPLACE FUNCTION fill_seq_invoice_credit_note() RETURNS TRIGGER
AS $$
	DECLARE
        resourcePath text;
    BEGIN
		resourcePath := current_setting('permission.resource_path', 't');
        SELECT coalesce(max(credit_note_sequence_number),0)+1 into NEW.credit_note_sequence_number from public.invoice_credit_note where resource_path = resourcePath;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_invoice_credit_note_seq BEFORE INSERT ON public.invoice_credit_note FOR EACH ROW EXECUTE PROCEDURE fill_seq_invoice_credit_note();