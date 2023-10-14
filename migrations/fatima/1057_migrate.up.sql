ALTER TABLE public.bill_item ADD COLUMN student_id text NOT NULL;

ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_student_id FOREIGN KEY(student_id) REFERENCES public.students(student_id);
