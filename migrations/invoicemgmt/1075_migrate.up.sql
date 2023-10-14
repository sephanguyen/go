ALTER TABLE IF EXISTS public.new_customer_code_history
     ALTER COLUMN bank_account_number TYPE TEXT USING bank_account_number::text;