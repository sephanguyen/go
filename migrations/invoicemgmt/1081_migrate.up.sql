ALTER TABLE IF EXISTS public.bulk_payment_validations_detail 
    ALTER COLUMN payment_status SET NOT NULL;