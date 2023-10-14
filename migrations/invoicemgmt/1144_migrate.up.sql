CREATE INDEX IF NOT EXISTS bill_item_student_id__idx ON public.bill_item(student_id);
CREATE INDEX IF NOT EXISTS invoice_adjustment_student_id__idx ON public.invoice_adjustment(student_id);
CREATE INDEX IF NOT EXISTS payment__created_at__idx_desc ON public.payment (created_at desc);
