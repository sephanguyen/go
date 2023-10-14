ALTER TABLE public.student_product ADD COLUMN product_associations jsonb;
ALTER TABLE public.order_item ADD COLUMN product_associations jsonb;