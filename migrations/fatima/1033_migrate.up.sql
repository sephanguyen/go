ALTER TABLE bill_item DROP CONSTRAINT bill_item_product_pk;
ALTER TABLE bill_item DROP CONSTRAINT fk_bill_item_bill_id;
ALTER TABLE bill_item DROP COLUMN bill_id;

ALTER TABLE bill_item ADD COLUMN order_id text NOT NULL;
ALTER TABLE bill_item ADD COLUMN bill_type text NOT NULL; -- billed at order, upcoming billing
ALTER TABLE bill_item ADD COLUMN billing_status text NOT NULL;  -- pending, billed
ALTER TABLE bill_item ADD COLUMN billing_date timestamp with time zone;
ALTER TABLE bill_item ADD COLUMN billing_from timestamp with time zone;
ALTER TABLE bill_item ADD COLUMN billing_to timestamp with time zone;
ALTER TABLE bill_item ADD COLUMN billing_schedule_period_id integer;
ALTER TABLE bill_item ADD COLUMN bill_item_sequence_number int NOT NULL;

ALTER TABLE ONLY public.bill_item ADD CONSTRAINT bill_item_order_bill_item_sequence_number_pk PRIMARY KEY (order_id, bill_item_sequence_number);
ALTER TABLE ONLY public.bill_item ADD CONSTRAINT fk_bill_item_billing_schedule_period_id FOREIGN KEY(billing_schedule_period_id) REFERENCES public.billing_schedule_period(id);
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_order_id FOREIGN KEY(order_id) REFERENCES public.order(id);
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_tax_id FOREIGN KEY(tax_id) REFERENCES public.tax(id);
ALTER TABLE public.order ADD CONSTRAINT fk_order_location_id FOREIGN KEY(location_id) REFERENCES public.location(id);

CREATE OR REPLACE FUNCTION fill_seq_bill_item() RETURNS TRIGGER
AS $$
	DECLARE
resourcePath text;
BEGIN
		resourcePath := current_setting('permission.resource_path', 't');
SELECT coalesce(max(bill_item_sequence_number),0)+1 into NEW.bill_item_sequence_number from public.bill_item where resource_path = resourcePath;
RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_bill_item_seq BEFORE INSERT ON public.bill_item FOR EACH ROW EXECUTE PROCEDURE fill_seq_bill_item();

ALTER TABLE ONLY public.bill_item
    ADD CONSTRAINT bill_item_sequence_number_resource_path_unique UNIQUE (bill_item_sequence_number,resource_path);

DROP TABLE IF EXISTS bill;
DROP function IF EXISTS fill_seq_bill();

DROP POLICY IF EXISTS rls_bill ON bill_item ;

CREATE POLICY rls_bill_item ON "bill_item" using (permission_check(resource_path, 'bill_item')) with check (permission_check(resource_path, 'bill_item'));

ALTER TABLE "bill_item" ENABLE ROW LEVEL security;
ALTER TABLE "bill_item" FORCE ROW LEVEL security;