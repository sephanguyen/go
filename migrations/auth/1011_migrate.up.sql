ALTER TABLE public.organizations
    ADD COLUMN IF NOT EXISTS salesforce_client_id text;
