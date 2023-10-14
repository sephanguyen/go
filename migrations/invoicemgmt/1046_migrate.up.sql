CREATE TABLE IF NOT EXISTS public.partner_convenience_store (
    partner_convenience_store_id text NOT NULL,
    manufacturer_code int NOT NULL,
    company_code int NOT NULL,
    shop_code text,
    company_name text NOT NULL,
    company_tel_number text,
    postal_code text,
    address1 text,
    address2 text,
    message1 text,
    message2 text,
    message3 text,
    message4 text,
    message5 text,
    message6 text,
    message7 text,
    message8 text,
    message9 text,
    message10 text,
    message11 text,
    message12 text,
    message13 text,
    message14 text,
    message15 text,
    message16 text,
    message17 text,
    message18 text,
    message19 text,
    message20 text,
    message21 text,
    message22 text,
    message23 text,
    message24 text,
    remarks text,
    is_archived BOOLEAN DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT partner_convenience_store___pk PRIMARY KEY (partner_convenience_store_id)
);

CREATE POLICY rls_partner_convenience_store ON "partner_convenience_store"
USING (permission_check(resource_path, 'partner_convenience_store'))
WITH CHECK (permission_check(resource_path, 'partner_convenience_store'));

CREATE POLICY rls_partner_convenience_store_restrictive ON "partner_convenience_store" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'partner_convenience_store'))
WITH CHECK (permission_check(resource_path, 'partner_convenience_store'));

ALTER TABLE "partner_convenience_store" ENABLE ROW LEVEL security;
ALTER TABLE "partner_convenience_store" FORCE ROW LEVEL security;
