ALTER TABLE ONLY public.partner_bank DROP COLUMN consigner_code;
ALTER TABLE ONLY public.partner_bank DROP COLUMN consigner_name;

ALTER TABLE IF EXISTS public.partner_bank
    ADD COLUMN IF NOT EXISTS consignor_code TEXT NOT NULL,
    ADD COLUMN IF NOT EXISTS consignor_name TEXT NOT NULL,
    ADD COLUMN IF NOT EXISTS is_default BOOLEAN DEFAULT false,
    ALTER COLUMN bank_number TYPE TEXT USING bank_number::text,
    ALTER COLUMN bank_branch_number TYPE TEXT USING bank_branch_number::text,
    ALTER COLUMN account_number TYPE TEXT USING account_number::text;