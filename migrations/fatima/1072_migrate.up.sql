ALTER TABLE public.student_product ADD COLUMN updated_from_student_product_id TEXT DEFAULT NULL;
ALTER TABLE public.student_product ADD COLUMN updated_to_student_product_id TEXT DEFAULT NULL;
ALTER TABLE public.student_product ADD COLUMN student_product_label TEXT DEFAULT NULL;
ALTER TABLE public.student_product ADD CONSTRAINT fk_updated_from_student_product_id FOREIGN KEY(updated_from_student_product_id) REFERENCES public.student_product(student_product_id);
ALTER TABLE public.student_product ADD CONSTRAINT fk_updated_to_student_product_id FOREIGN KEY(updated_to_student_product_id) REFERENCES public.student_product(student_product_id);
