ALTER TABLE public.product_setting ADD COLUMN IF NOT EXISTS is_operation_fee boolean DEFAULT false;
