ALTER TABLE public.student_product ALTER COLUMN start_date DROP NOT NULL;
ALTER TABLE public.student_product ALTER COLUMN end_date DROP NOT NULL;
ALTER TABLE public.student_product ALTER COLUMN product_status SET NOT NULL;
