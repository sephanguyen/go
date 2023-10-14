ALTER TABLE IF EXISTS public.invoice
  DROP COLUMN IF EXISTS is_expired;

ALTER TABLE IF EXISTS public.payment
  DROP COLUMN IF EXISTS is_expired;