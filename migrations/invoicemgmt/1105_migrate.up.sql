ALTER TABLE public.payment 
    ADD COLUMN IF NOT EXISTS bulk_payment_id text;