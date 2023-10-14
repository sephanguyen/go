ALTER TABLE IF EXISTS public.bank_mapping
    ALTER created_at SET DEFAULT now();
ALTER TABLE IF EXISTS public.bank_mapping
    ALTER updated_at SET DEFAULT now();
ALTER TABLE IF EXISTS public.bank_mapping
    ALTER deleted_at SET DEFAULT NULL;
