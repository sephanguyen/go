CREATE TABLE public.upcoming_bill_item (
                                           order_id text NOT NULL,
                                           bill_item_sequence_number int,
                                           product_id text NOT NULL,
                                           student_product_id text,
                                           product_description text NOT NULL,
                                           discount_id text,
                                           tax_id text NOT NULL,
                                           billing_schedule_period_id text,
                                           billing_date timestamp with time zone,
                                           created_at timestamp with time zone NOT NULL,
                                           updated_at timestamp with time zone NOT NULL,
                                           deleted_at timestamp with time zone,
                                           is_generated bool DEFAULT FALSE,
                                           execute_note text,
                                           resource_path text DEFAULT autofillresourcepath()
);
CREATE POLICY rls_upcoming_bill_item ON "upcoming_bill_item"
    using (permission_check(resource_path, 'upcoming_bill_item'))
    with check (permission_check(resource_path, 'upcoming_bill_item'));

CREATE POLICY rls_upcoming_bill_item_restrictive ON "upcoming_bill_item"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'upcoming_bill_item'))
    WITH CHECK (permission_check(resource_path, 'upcoming_bill_item'));

ALTER TABLE "upcoming_bill_item" ENABLE ROW LEVEL security;
ALTER TABLE "upcoming_bill_item" FORCE ROW LEVEL security;
