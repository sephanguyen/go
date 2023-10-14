ALTER TABLE public.payment 
    ADD COLUMN IF NOT EXISTS migrated_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS payment_reference_id TEXT;

ALTER TABLE public.invoice 
    ADD COLUMN IF NOT EXISTS migrated_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS invoice_reference_id TEXT,
    ADD COLUMN IF NOT EXISTS invoice_reference_id2 TEXT;

ALTER TABLE public.invoice_bill_item
    ADD COLUMN IF NOT EXISTS migrated_at TIMESTAMP WITH TIME ZONE;