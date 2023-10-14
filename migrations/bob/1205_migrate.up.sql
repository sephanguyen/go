ALTER TABLE
    public.organizations
    ADD COLUMN IF NOT EXISTS scrypt_signer_key text;
ALTER TABLE
    public.organizations
    ADD COLUMN IF NOT EXISTS scrypt_salt_separator text;
ALTER TABLE
    public.organizations
    ADD COLUMN IF NOT EXISTS scrypt_rounds text;
ALTER TABLE
    public.organizations
    ADD COLUMN IF NOT EXISTS scrypt_memory_cost text;
