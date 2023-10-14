ALTER TABLE IF EXISTS public.organizations
    ADD COLUMN IF NOT EXISTS name text,
    ADD COLUMN IF NOT EXISTS logo_url text;
