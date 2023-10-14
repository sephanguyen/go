-- =========================================================================================
-- ============================== public.bank_account table ==============================
-- =========================================================================================
CREATE TABLE IF NOT EXISTS public.bank_account (
    bank_account_id text NOT NULL,
    student_payment_detail_id text NOT NULL,
    student_id text NOT NULL,
    is_verified BOOLEAN DEFAULT false,
    bank_branch_id text NOT NULL,
    bank_account_number text NOT NULL,
    bank_account_holder text NOT NULL,
    bank_account_type text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    bank_id text,

    CONSTRAINT bank_account__pk PRIMARY KEY (bank_account_id)
);

CREATE POLICY rls_bank_account ON "bank_account"
USING (permission_check(resource_path, 'bank_account'))
WITH CHECK (permission_check(resource_path, 'bank_account'));

CREATE POLICY rls_bank_account_restrictive ON "bank_account" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'bank_account'))
WITH CHECK (permission_check(resource_path, 'bank_account'));

ALTER TABLE "bank_account" ENABLE ROW LEVEL security;
ALTER TABLE "bank_account" FORCE ROW LEVEL security;

-- =========================================================================================
-- ============================== public.bank table ==============================
-- =========================================================================================
CREATE TABLE IF NOT EXISTS public.bank (
	bank_id text NOT NULL,
	bank_code text NOT NULL,
	bank_name text NOT NULL,
	bank_name_phonetic text NOT NULL,
	is_archived bool NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT bank__pk PRIMARY KEY (bank_id)
);

ALTER TABLE ONLY public.bank
	DROP CONSTRAINT IF EXISTS bank__bank_code__unique;

DROP POLICY IF EXISTS rls_bank on "bank";
DROP POLICY IF EXISTS rls_bank_restrictive on "bank";

CREATE POLICY rls_bank ON "bank"
USING (permission_check(resource_path, 'bank'))
WITH CHECK (permission_check(resource_path, 'bank'));

CREATE POLICY rls_bank_restrictive ON "bank" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'bank'))
WITH CHECK (permission_check(resource_path, 'bank'));

ALTER TABLE "bank" ENABLE ROW LEVEL security;
ALTER TABLE "bank" FORCE ROW LEVEL security;
