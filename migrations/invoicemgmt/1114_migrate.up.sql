ALTER TABLE IF EXISTS ONLY public.invoice ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE IF EXISTS ONLY public.payment ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE IF EXISTS ONLY public.invoice_bill_item ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE IF EXISTS ONLY public.invoice_bill_item ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone;