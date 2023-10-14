CREATE TABLE public.bill (
    id text NOT NULL,
    bill_sequence_number int NOT NULL,
    order_id text NOT NULL,
    bill_type text NOT NULL, -- billed at order, upcoming billing
    total_amount int,
    billing_status text NOT NULL, -- pending, billed
    billing_date timestamp with time zone,
    billing_from timestamp with time zone,
    billing_to timestamp with time zone,
    billing_schedule_period_id integer,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.bill ADD CONSTRAINT bill_pk PRIMARY KEY (id);
ALTER TABLE public.bill ADD CONSTRAINT fk_bill_order_id FOREIGN KEY(order_id) REFERENCES public.order(id);

CREATE POLICY rls_bill ON "bill" using (permission_check(resource_path, 'bill')) with check (permission_check(resource_path, 'bill'));

ALTER TABLE "bill" ENABLE ROW LEVEL security;
ALTER TABLE "bill" FORCE ROW LEVEL security;


CREATE TABLE public.bill_item (
    bill_id text NOT NULL,
    product_id integer NOT NULL,
    product_description text NOT NULL,
    product_pricing integer NOT NULL,
    discount_amount_type text,
    discount_amount_value numeric(12,2) NOT NULL,
    tax_id integer NOT NULL,
    tax_category text NOT NULL,
    tax_percentage integer NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.bill_item ADD CONSTRAINT bill_item_product_pk PRIMARY KEY (bill_id, product_id);
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_bill_id FOREIGN KEY(bill_id) REFERENCES public.bill(id);
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_product_id FOREIGN KEY(product_id) REFERENCES public.product(id);

CREATE POLICY rls_bill ON "bill_item" using (permission_check(resource_path, 'bill_item')) with check (permission_check(resource_path, 'bill_item'));

ALTER TABLE "bill_item" ENABLE ROW LEVEL security;
ALTER TABLE "bill_item" FORCE ROW LEVEL security;



CREATE OR REPLACE FUNCTION fill_seq_bill() RETURNS TRIGGER
AS $$
	DECLARE
        resourcePath text;
    BEGIN
		resourcePath := current_setting('permission.resource_path', 't');
        SELECT coalesce(max(bill_sequence_number),0)+1 into NEW.bill_sequence_number from public.bill where resource_path = resourcePath;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_bill_seq BEFORE INSERT ON public.bill FOR EACH ROW EXECUTE PROCEDURE fill_seq_bill();

ALTER TABLE ONLY public.bill
    ADD CONSTRAINT bill_sequence_number_resource_path_unique UNIQUE (bill_sequence_number,resource_path);
