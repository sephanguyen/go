CREATE TABLE IF NOT EXISTS public.partner_bank (
    partner_bank_id TEXT NOT NULL,
    consigner_code INTEGER NOT NULL,
    consigner_name TEXT NOT NULL,
    bank_number INTEGER NOT NULL,
    bank_name TEXT NOT NULL,
    bank_branch_number INTEGER NOT NULL,
    bank_branch_name TEXT NOT NULL,
    deposit_items TEXT NOT NULL,
    account_number INTEGER NOT NULL,
    remarks TEXT,
    is_archived BOOLEAN DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT partner_bank__pk PRIMARY KEY (partner_bank_id)
);

CREATE POLICY rls_partner_bank ON "partner_bank"
USING (permission_check(resource_path, 'partner_bank'))
WITH CHECK (permission_check(resource_path, 'partner_bank'));

CREATE POLICY rls_partner_bank_restrictive ON "partner_bank" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'partner_bank'))
WITH CHECK (permission_check(resource_path, 'partner_bank'));

ALTER TABLE "partner_bank" ENABLE ROW LEVEL security;
ALTER TABLE "partner_bank" FORCE ROW LEVEL security;
