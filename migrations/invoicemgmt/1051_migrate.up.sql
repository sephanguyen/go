ALTER TABLE ONLY public.bill_item
  DROP CONSTRAINT IF EXISTS bill_item_student_id_fk;
ALTER TABLE ONLY public.bill_item
  DROP CONSTRAINT IF EXISTS fk_bill_item_location_id;