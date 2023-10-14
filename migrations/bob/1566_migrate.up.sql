ALTER TABLE public.student_product ADD CONSTRAINT student_product_product_fk FOREIGN KEY (product_id) REFERENCES "product"(product_id);
