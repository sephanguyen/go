ALTER TABLE public.student_associated_product
DROP CONSTRAINT IF EXISTS student_associated_product_pk;

ALTER TABLE ONLY public.student_associated_product
ADD CONSTRAINT student_associated_product_pk PRIMARY KEY (student_product_id, associated_product_id);
