ALTER TABLE ONLY public.bulk_payment_request_file
  DROP CONSTRAINT IF EXISTS bulk_payment_request_file_sequence_bulk_payment_request_unique;

CREATE INDEX IF NOT EXISTS student_payment_detail_student_id__idx ON public.student_payment_detail(student_id);
CREATE INDEX IF NOT EXISTS bank_account_student_id__idx ON public.bank_account(student_id);
CREATE INDEX IF NOT EXISTS billing_address_student_id__idx ON public.billing_address(user_id);
CREATE INDEX IF NOT EXISTS new_customer_code_history_student_id__idx ON public.new_customer_code_history(student_id);