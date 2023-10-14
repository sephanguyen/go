CREATE OR REPLACE FUNCTION fill_seq_invoice() RETURNS TRIGGER
AS $$
	DECLARE
        resourcePath text;
    BEGIN
		resourcePath := current_setting('permission.resource_path', 't');
        SELECT coalesce(max(invoice_id),0)+1 into NEW.invoice_id from public.invoice where resource_path = resourcePath;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_invoice_seq BEFORE INSERT ON public.invoice FOR EACH ROW EXECUTE PROCEDURE fill_seq_invoice();

ALTER TABLE ONLY public.invoice
    ADD CONSTRAINT invoice_id_resource_path_unique UNIQUE (invoice_id,resource_path);