ALTER TABLE public.invoice_schedule_student ADD COLUMN IF NOT EXISTS actual_error_details TEXT;
ALTER TABLE public.invoice_schedule_history DROP COLUMN IF EXISTS number_of_students_without_bill_items;