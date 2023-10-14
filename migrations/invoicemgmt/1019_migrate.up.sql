-- Payment table changes

ALTER TABLE ONLY public.payment 
    ALTER COLUMN payment_id TYPE TEXT;             -- prevents conflict with multiple DB per partner setup
ALTER TABLE ONLY public.payment 
    ALTER COLUMN payment_id DROP DEFAULT;          -- removes sequencing via nextval
ALTER TABLE ONLY public.payment 
    ADD COLUMN result text;
ALTER TABLE ONLY public.payment 
    ADD COLUMN payment_sequence_number int;  

DROP SEQUENCE IF EXISTS public.payment_id_seq;     -- payment_id will be text; sequence not needed anymore

CREATE SEQUENCE public.payment_sequence_number_seq 
    AS integer;
ALTER SEQUENCE public.payment_sequence_number_seq OWNED BY public.payment.payment_sequence_number;

ALTER TABLE ONLY public.payment 
    ALTER COLUMN payment_sequence_number SET DEFAULT nextval('public.payment_sequence_number_seq'::regclass);
ALTER TABLE ONLY public.payment
    ADD CONSTRAINT payment_sequence_number_resource_path_unique UNIQUE (payment_sequence_number, resource_path);

CREATE OR REPLACE FUNCTION fill_seq_payment() RETURNS TRIGGER
AS $$
	DECLARE
        resourcePath text;
    BEGIN
		resourcePath := current_setting('permission.resource_path', 't');
        SELECT coalesce(max(payment_sequence_number),0)+1 into NEW.payment_sequence_number from public.payment where resource_path = resourcePath;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_payment_seq BEFORE INSERT ON public.payment FOR EACH ROW EXECUTE PROCEDURE fill_seq_payment();

-- Invoice table changes
-- Temporarily drop all FK constraints from other tables related to invoice_id; will be recreated later
ALTER TABLE public.payment 
    DROP CONSTRAINT payment_invoice_fk;
ALTER TABLE public.invoice_action_log 
    DROP CONSTRAINT invoice_action_log_invoice_fk;
ALTER TABLE public.invoice_bill_item 
    DROP CONSTRAINT invoice_bill_item_invoice_fk;

-- Change invoice_id FK from the other tables
ALTER TABLE ONLY public.invoice_bill_item
    ALTER COLUMN invoice_id TYPE TEXT;
ALTER TABLE ONLY public.payment
    ALTER COLUMN invoice_id TYPE TEXT;
ALTER TABLE ONLY public.invoice_action_log
    ALTER COLUMN invoice_id TYPE TEXT;
ALTER TABLE ONLY public.invoice_bill_item
    ALTER COLUMN invoice_id TYPE TEXT;

CREATE SEQUENCE public.invoice_sequence_number_seq 
    AS integer;

ALTER TABLE ONLY public.invoice
    ADD COLUMN invoice_sequence_number int;  
ALTER TABLE ONLY public.invoice
    ALTER COLUMN invoice_sequence_number SET DEFAULT nextval('public.invoice_sequence_number_seq'::regclass);
ALTER TABLE ONLY public.invoice
    ALTER COLUMN invoice_id TYPE TEXT;       -- prevents conflict with multiple DB per partner setup
ALTER TABLE ONLY public.invoice
    ADD CONSTRAINT invoice_sequence_number_resource_path_unique UNIQUE (invoice_sequence_number,resource_path);
ALTER TABLE public.invoice 
    DROP CONSTRAINT invoice_id_resource_path_unique;

ALTER SEQUENCE public.invoice_sequence_number_seq OWNED 
    BY public.invoice.invoice_sequence_number;

CREATE OR REPLACE FUNCTION fill_seq_invoice() RETURNS TRIGGER
AS $$
	DECLARE
        resourcePath text;
    BEGIN
		resourcePath := current_setting('permission.resource_path', 't');
        SELECT coalesce(max(invoice_sequence_number),0)+1 into NEW.invoice_sequence_number from public.invoice where resource_path = resourcePath;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

-- Recreate FK constraints
ALTER TABLE public.invoice_bill_item 
    ADD CONSTRAINT invoice_bill_item_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice" (invoice_id);
ALTER TABLE public.payment 
    ADD CONSTRAINT payment_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice" (invoice_id);
ALTER TABLE public.invoice_action_log 
    ADD CONSTRAINT invoice_action_log_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice" (invoice_id);