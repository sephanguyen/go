ALTER TABLE public.invoice 
    ADD COLUMN IF NOT EXISTS outstanding_balance numeric(12,2),
    ADD COLUMN IF NOT EXISTS amount_paid numeric(12,2),
    ADD COLUMN IF NOT EXISTS amount_refunded numeric(12,2);