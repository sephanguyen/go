-- Update bill_item schema because order schema change master data id from integer to uuid
ALTER TABLE public.bill_item ALTER COLUMN tax_id TYPE text;
ALTER TABLE public.bill_item ALTER COLUMN billing_schedule_period_id TYPE text;
