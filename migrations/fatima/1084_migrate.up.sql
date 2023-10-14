ALTER TABLE public.order_item
    ADD COLUMN IF NOT EXISTS effective_date timestamptz NULL;
ALTER TABLE public.order_item
    ADD COLUMN IF NOT EXISTS cancellation_date timestamptz NULL;