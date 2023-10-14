ALTER TABLE public.invoice_schedule ADD COLUMN IF NOT EXISTS remarks TEXT;
ALTER TABLE public.invoice_schedule ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT false;