CREATE OR REPLACE FUNCTION fill_seq_order() RETURNS TRIGGER
AS $$
	DECLARE
        resourcePath text;
    BEGIN
		resourcePath := current_setting('permission.resource_path', 't');
        SELECT coalesce(max(order_sequence_number),0)+1 into NEW.order_sequence_number from public.order where resource_path = resourcePath;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_order_seq BEFORE INSERT ON public.order FOR EACH ROW EXECUTE PROCEDURE fill_seq_order();

ALTER TABLE ONLY public.order
    ADD CONSTRAINT order_sequence_number_resource_path_unique UNIQUE (order_sequence_number,resource_path);