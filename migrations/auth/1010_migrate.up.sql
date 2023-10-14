ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS "username"    TEXT,
    ADD COLUMN IF NOT EXISTS "login_email" TEXT;

DROP INDEX IF EXISTS users__username__idx;
CREATE INDEX users__username__idx ON users (username, resource_path);
