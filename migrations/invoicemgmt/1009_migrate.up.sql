ALTER TABLE ONLY public.bill_item ADD CONSTRAINT bill_item_sequence_number_resource_path_unique UNIQUE (bill_item_sequence_number,resource_path);

ALTER TABLE ONLY public.invoice_billing_item ADD CONSTRAINT invoice_bill_item_fk FOREIGN KEY(bill_item_sequence_number, resource_path) REFERENCES public.bill_item(bill_item_sequence_number, resource_path);

ALTER TABLE ONLY public.bill_item ADD CONSTRAINT bill_item_student_id_fk FOREIGN KEY(student_id) REFERENCES public.students(student_id);