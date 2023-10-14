ALTER TABLE IF EXISTS public.organizations
    ADD COLUMN IF NOT EXISTS domain_name text,
    ADD COLUMN IF NOT EXISTS logo_url text,
    ADD COLUMN IF NOT EXISTS country text,
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
