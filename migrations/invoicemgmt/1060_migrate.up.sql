ALTER TABLE IF EXISTS public.partner_convenience_store
    ALTER created_at SET DEFAULT now();
ALTER TABLE IF EXISTS public.partner_convenience_store
    ALTER updated_at SET DEFAULT now();
ALTER TABLE IF EXISTS public.partner_convenience_store
    ALTER deleted_at SET DEFAULT NULL;
