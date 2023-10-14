CREATE TABLE IF NOT EXISTS public.invoice_adjustment (
    invoice_adjustment_id TEXT NOT NULL,
    invoice_id TEXT NOT NULL,
    description TEXT NOT NULL,
    amount numeric(12,2) NOT NULL,
    student_id TEXT NOT NULL,
    invoice_adjustment_sequence_number INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__invoice_adjustment PRIMARY KEY (invoice_adjustment_id),
    CONSTRAINT invoice_adjustment_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice"(invoice_id),
    CONSTRAINT invoice_adjustment_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id)
);


CREATE SEQUENCE public.invoice_adjustment_sequence_number_seq 
    AS integer;
ALTER SEQUENCE public.invoice_adjustment_sequence_number_seq OWNED BY public.invoice_adjustment.invoice_adjustment_sequence_number;

ALTER TABLE ONLY public.invoice_adjustment 
    ALTER COLUMN invoice_adjustment_sequence_number SET DEFAULT nextval('public.invoice_adjustment_sequence_number_seq'::regclass);
ALTER TABLE ONLY public.invoice_adjustment
    ADD CONSTRAINT invoice_adjustment_sequence_number_resource_path_unique UNIQUE (invoice_adjustment_sequence_number, resource_path);

CREATE OR REPLACE FUNCTION fill_seq_invoice_adjustment() RETURNS TRIGGER
AS $$
	DECLARE
        resourcePath text;
    BEGIN
		resourcePath := current_setting('permission.resource_path', 't');
        SELECT coalesce(max(invoice_adjustment_sequence_number),0)+1 into NEW.invoice_adjustment_sequence_number from public.invoice_adjustment where resource_path = resourcePath;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_invoice_adjustment_seq BEFORE INSERT ON public.invoice_adjustment FOR EACH ROW EXECUTE PROCEDURE fill_seq_invoice_adjustment();


CREATE POLICY rls_invoice_adjustment ON "invoice_adjustment"
USING (permission_check(resource_path, 'invoice_adjustment')) WITH CHECK (permission_check(resource_path, 'invoice_adjustment'));

CREATE POLICY rls_invoice_adjustment_restrictive ON "invoice_adjustment" AS RESTRICTIVE
USING (permission_check(resource_path, 'invoice_adjustment'))WITH CHECK (permission_check(resource_path, 'invoice_adjustment'));

ALTER TABLE "invoice_adjustment" ENABLE ROW LEVEL security;
ALTER TABLE "invoice_adjustment" FORCE ROW LEVEL security;
