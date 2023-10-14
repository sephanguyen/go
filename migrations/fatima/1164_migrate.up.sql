ALTER TABLE public.product_group ADD COLUMN IF NOT EXISTS discount_type text;
ALTER TABLE public.product_group ADD COLUMN IF NOT EXISTS is_archived boolean NOT NULL DEFAULT false;

ALTER TABLE public.student_discount_tracker ADD COLUMN IF NOT EXISTS product_group_id text;
ALTER TABLE public.student_discount_tracker
    ADD CONSTRAINT fk_student_discount_tracker_product_group_id FOREIGN KEY (product_group_id) REFERENCES public.product_group(product_group_id);
