ALTER TABLE IF EXISTS public.bank_account
    ADD COLUMN IF NOT EXISTS bank_id text;

ALTER TABLE IF EXISTS public.bank_account
    DROP CONSTRAINT IF EXISTS bank_account__bank_id__fk;
ALTER TABLE IF EXISTS public.bank_account
    ADD CONSTRAINT bank_account__bank_id__fk FOREIGN KEY (bank_id) REFERENCES bank (bank_id);
