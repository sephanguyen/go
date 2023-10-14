-- Add unique payment_id constraint
ALTER TABLE public.bulk_payment_request_file_payment ADD CONSTRAINT bulk_payment_request_file_payment_payment_id_key UNIQUE(payment_id);

-- Remove total_file_count from bulk_payment_request
ALTER TABLE public.bulk_payment_request DROP COLUMN IF EXISTS total_file_count;

-- Drop the not null constraint of payment due dates
ALTER TABLE public.bulk_payment_request ALTER payment_due_date_from DROP NOT NULL;
ALTER TABLE public.bulk_payment_request ALTER payment_due_date_to DROP NOT NULL;

-- Add total_file_count and parent file to bulk_payment_request_file
ALTER TABLE public.bulk_payment_request_file ADD COLUMN IF NOT EXISTS total_file_count INTEGER;
ALTER TABLE public.bulk_payment_request_file ADD COLUMN IF NOT EXISTS parent_payment_request_file_id TEXT NULL;

ALTER TABLE public.bulk_payment_request_file ALTER file_url DROP NOT NULL;