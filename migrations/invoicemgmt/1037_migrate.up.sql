CREATE TABLE IF NOT EXISTS public.discount (
    discount_id text NOT NULL,
    name text NOT NULL,
    discount_type text NOT NULL,
    discount_amount_type text NOT NULL,
    discount_amount_value numeric(12,2) NOT NULL,
    recurring_valid_duration integer NULL,
    available_from timestamp with time zone NOT NULL,
    available_until timestamp with time zone NOT NULL,
    remarks text NULL,
    is_archived boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.discount
    ADD CONSTRAINT discount_pk PRIMARY KEY (discount_id);

CREATE POLICY rls_discount ON "discount" using (permission_check(resource_path, 'discount')) with check (permission_check(resource_path, 'discount'));

CREATE POLICY rls_discount_restrictive ON "discount" 
    AS RESTRICTIVE TO public 
    USING (permission_check(resource_path, 'discount'))
    WITH CHECK (permission_check(resource_path, 'discount'));

ALTER TABLE "discount" ENABLE ROW LEVEL security;
ALTER TABLE "discount" FORCE ROW LEVEL security;


ALTER TABLE ONLY public.bill_item
    ADD COLUMN IF NOT EXISTS discount_id text;