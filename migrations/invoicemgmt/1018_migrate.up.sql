ALTER TABLE ONLY public.bill_item
  DROP CONSTRAINT IF EXISTS fk_bill_item_student_product_id;

ALTER TABLE public.bill_item
  DROP COLUMN IF EXISTS student_product_id;

DROP TABLE IF EXISTS public.student_product;