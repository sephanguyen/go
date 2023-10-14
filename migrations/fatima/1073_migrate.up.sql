CREATE TABLE public.order_item_custom (
    order_id text NOT NULL,
    name text NOT NULL,
    tax_id integer,
    price numeric(12,2) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.order_item_custom ADD CONSTRAINT order_item_custom_pk PRIMARY KEY (order_id,name);
ALTER TABLE public.order_item_custom ADD CONSTRAINT fk_order_item_custom_order_id FOREIGN KEY(order_id) REFERENCES public.order(order_id);
ALTER TABLE public.order_item_custom ADD CONSTRAINT fk_order_item_custom_tax_id FOREIGN KEY(tax_id) REFERENCES public.tax(tax_id);

CREATE POLICY rls_order_item_custom ON "order_item_custom" using (permission_check(resource_path, 'order_item_custom')) with check (permission_check(resource_path, 'order_item_custom'));

ALTER TABLE "order_item_custom" ENABLE ROW LEVEL security;
ALTER TABLE "order_item_custom" FORCE ROW LEVEL security;

CREATE TABLE public.bill_item_custom (
    bill_item_sequence_number int NOT NULL,
    order_id text NOT NULL,
    student_id text NOT NULL,
    bill_type text NOT NULL,
    billing_status text NOT NULL,
    billing_date timestamp with time zone,
    billing_from timestamp with time zone,
    billing_to timestamp with time zone,
    product_name text NOT NULL,
    product_price numeric(12,2) NOT NULL,
    tax_id integer,
    tax_category text,
    tax_amount numeric(12,2),
    tax_percentage integer,
    location_id text NOT NULL,
    location_name text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    billing_approval_status text,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.bill_item_custom ADD CONSTRAINT bill_item_custom_pk PRIMARY KEY (bill_item_sequence_number, order_id);
ALTER TABLE ONLY public.bill_item_custom ADD CONSTRAINT bill_item_custom_order_id_fk FOREIGN KEY (order_id) REFERENCES public.order(order_id);
ALTER TABLE ONLY public.bill_item_custom ADD CONSTRAINT bill_item_custom_student_id_fk FOREIGN KEY(student_id) REFERENCES public.students(student_id);
ALTER TABLE ONLY public.bill_item_custom ADD CONSTRAINT bill_item_custom_tax_id_fk FOREIGN KEY(tax_id) REFERENCES public.tax(tax_id);
ALTER TABLE ONLY public.bill_item_custom ADD CONSTRAINT bill_item_custom_location_id_fk FOREIGN KEY (location_id) REFERENCES public.locations(location_id);

CREATE POLICY rls_bill_item_custom ON "bill_item_custom" using (permission_check(resource_path, 'bill_item_custom')) with check (permission_check(resource_path, 'bill_item_custom'));

ALTER TABLE "bill_item_custom" ENABLE ROW LEVEL security;
ALTER TABLE "bill_item_custom" FORCE ROW LEVEL security;

CREATE OR REPLACE FUNCTION fill_seq_bill_item_custom() RETURNS TRIGGER
AS $$
    DECLARE
resourcePath text;
BEGIN
        resourcePath := current_setting('permission.resource_path', 't');
SELECT coalesce(max(bill_item_sequence_number),0)+1 into NEW.bill_item_sequence_number from public.bill_item_custom where resource_path = resourcePath;
RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER fill_in_bill_item_seq BEFORE INSERT ON public.bill_item_custom FOR EACH ROW EXECUTE PROCEDURE fill_seq_bill_item_custom();

ALTER TABLE ONLY public.bill_item_custom
    ADD CONSTRAINT bill_item_custom_sequence_number_resource_path_unique UNIQUE (bill_item_sequence_number, resource_path);
