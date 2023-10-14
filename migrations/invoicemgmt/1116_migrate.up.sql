ALTER TABLE public.student_payment_detail 
    ADD COLUMN IF NOT EXISTS migrated_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE public.bank_account 
    ADD COLUMN IF NOT EXISTS migrated_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE public.billing_address 
    ADD COLUMN IF NOT EXISTS migrated_at TIMESTAMP WITH TIME ZONE;