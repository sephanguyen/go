ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS "email" TEXT;

DROP INDEX IF EXISTS users__email__idx;
CREATE INDEX users__email__idx ON users (email, resource_path);
