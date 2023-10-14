ALTER TABLE public.invoice_adjustment
    ADD COLUMN IF NOT EXISTS migrated_at TIMESTAMP WITH TIME ZONE;