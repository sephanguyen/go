ALTER TABLE public.discount ALTER order_item_updated_at DROP DEFAULT;
ALTER TABLE public.discount ALTER COLUMN order_item_updated_at DROP NOT NULL;
