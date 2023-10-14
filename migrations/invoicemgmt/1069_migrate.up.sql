ALTER TABLE public.bulk_payment_validations ADD COLUMN IF NOT EXISTS pending_payments int NOT NULL DEFAULT 0;
ALTER TABLE IF EXISTS public.bulk_payment_validations RENAME COLUMN successful_validations TO successful_payments;
ALTER TABLE IF EXISTS public.bulk_payment_validations ALTER COLUMN successful_payments SET DEFAULT 0;
ALTER TABLE IF EXISTS public.bulk_payment_validations RENAME COLUMN failed_validations TO failed_payments;
ALTER TABLE IF EXISTS public.bulk_payment_validations ALTER COLUMN failed_payments SET DEFAULT 0;