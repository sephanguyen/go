ALTER TABLE public.product_setting ADD COLUMN IF NOT EXISTS is_pausable boolean DEFAULT true;
