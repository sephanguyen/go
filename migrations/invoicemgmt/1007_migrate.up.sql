CREATE TABLE public.bill_item (
    product_id integer NOT NULL,
    product_description text NOT NULL,
    product_pricing integer NOT NULL,
    discount_amount_type text,
    discount_amount_value numeric(12,2) NOT NULL,
    tax_id integer,
    tax_category text,
    tax_percentage integer,
    resource_path text DEFAULT autofillresourcepath(),
    order_id text NOT NULL,
    bill_type text NOT NULL, -- billed at order, upcoming billing
    billing_status text NOT NULL,  -- pending, billed
    billing_date timestamp with time zone,
    billing_from timestamp with time zone,
    billing_to timestamp with time zone,
    billing_schedule_period_id integer,
    bill_item_sequence_number int NOT NULL,
    discount_amount numeric(12,2) NOT NULL,
    tax_amount numeric(12,2),
    final_price numeric(12,2) NOT NULL,
    student_id text NOT NULL,
    student_product_id text NOT NULL,
    billing_approval_status text,
    billing_item_description JSONB,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);

ALTER TABLE ONLY public.bill_item ADD CONSTRAINT bill_item_product_pk PRIMARY KEY (order_id, bill_item_sequence_number);

CREATE POLICY rls_bill_item ON "bill_item" using (permission_check(resource_path, 'bill_item')) with check (permission_check(resource_path, 'bill_item'));

ALTER TABLE "bill_item" ENABLE ROW LEVEL security;
ALTER TABLE "bill_item" FORCE ROW LEVEL security;
