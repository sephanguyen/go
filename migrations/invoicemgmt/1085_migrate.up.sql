ALTER TABLE IF EXISTS public.billing_address 
    ADD COLUMN IF NOT EXISTS prefecture_code text;