ALTER TABLE public.payment 
  ADD COLUMN IF NOT EXISTS receipt_date timestamp with time zone;
