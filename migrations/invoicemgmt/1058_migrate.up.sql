ALTER TABLE IF EXISTS public.bank
    ALTER created_at SET DEFAULT now();
ALTER TABLE IF EXISTS public.bank
    ALTER updated_at SET DEFAULT now();
ALTER TABLE IF EXISTS public.bank
    ALTER deleted_at SET DEFAULT NULL;

ALTER TABLE IF EXISTS public.bank_branch
    ALTER created_at SET DEFAULT now();
ALTER TABLE IF EXISTS public.bank_branch
    ALTER updated_at SET DEFAULT now();
ALTER TABLE IF EXISTS public.bank_branch
    ALTER deleted_at SET DEFAULT NULL;
