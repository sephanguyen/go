ALTER TABLE bill_item DROP CONSTRAINT IF EXISTS fk_bill_item_student_id;
ALTER TABLE student_product DROP CONSTRAINT IF EXISTS fk_student_product_student_id;
ALTER TABLE bill_item ADD CONSTRAINT fk_bill_item_student_id FOREIGN KEY (student_id) REFERENCES students (student_id);
ALTER TABLE student_product ADD CONSTRAINT fk_student_product_student_id FOREIGN KEY (student_id) REFERENCES students (student_id);
