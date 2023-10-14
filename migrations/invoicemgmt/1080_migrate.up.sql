ALTER TABLE IF EXISTS public.bulk_payment_validations_detail 
    ADD COLUMN IF NOT EXISTS payment_status text;