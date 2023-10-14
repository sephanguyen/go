ALTER TABLE public.invoice_schedule 
  ADD COLUMN IF NOT EXISTS scheduled_date timestamp with time zone;
