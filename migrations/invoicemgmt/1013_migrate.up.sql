ALTER TABLE public.invoice ADD COLUMN IF NOT EXISTS sub_total numeric(12,2) NOT NULL;
ALTER TABLE public.invoice ADD COLUMN IF NOT EXISTS total numeric(12,2) NOT NULL;