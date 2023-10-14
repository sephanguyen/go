ALTER TABLE public.billing_address
  DROP COLUMN IF EXISTS prefecture_name,
  DROP COLUMN IF EXISTS prefecture_id;