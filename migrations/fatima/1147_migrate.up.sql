ALTER TABLE public.product_price ADD COLUMN IF NOT EXISTS price_type TEXT DEFAULT 'DEFAULT_PRICE';
ALTER TABLE ONLY public.product_price ADD CONSTRAINT product_price_type_check CHECK ((price_type = ANY('{DEFAULT_PRICE, ENROLLED_PRICE}'::text[])));

DROP INDEX IF EXISTS product_price_uni_idx;
CREATE UNIQUE INDEX product_price_uni_idx ON public.product_price (product_id, COALESCE(billing_schedule_period_id, ''), COALESCE(quantity, -1), COALESCE(price_type, ''));
