ALTER TABLE IF EXISTS public.invoice_action_log 
    ADD COLUMN IF NOT EXISTS bulk_payment_validations_id text;

ALTER TABLE IF EXISTS public.bulk_payment_validations_detail 
    ADD COLUMN IF NOT EXISTS previous_result_code text;