CREATE TABLE public.bill_item_account_category (
    bill_item_sequence_number int NOT NULL,
    accounting_category_id text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT bill_item_account_category_pk PRIMARY KEY (bill_item_sequence_number, accounting_category_id,resource_path),
    CONSTRAINT fk_bill_item_account_category_bill_item_sequence_number FOREIGN KEY (bill_item_sequence_number,resource_path) REFERENCES public.bill_item (bill_item_sequence_number,resource_path),
    CONSTRAINT fk_bill_item_account_category_accounting_category_id FOREIGN KEY (accounting_category_id) REFERENCES public.accounting_category(accounting_category_id)
);

CREATE POLICY rls_bill_item_account_category ON "bill_item_account_category"
    USING (permission_check(resource_path, 'bill_item_account_category'))
    WITH CHECK (permission_check(resource_path, 'bill_item_account_category'));

CREATE POLICY rls_bill_item_account_category_restrictive ON "bill_item_account_category"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'bill_item_account_category'))
    WITH CHECK (permission_check(resource_path, 'bill_item_account_category'));

ALTER TABLE "bill_item_account_category" ENABLE ROW LEVEL security;
ALTER TABLE "bill_item_account_category" FORCE ROW LEVEL security;
