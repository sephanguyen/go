ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS country text NOT NULL DEFAULT '';

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS avatar text;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS phone_number text;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS email text;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS device_token text;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS allow_notification boolean;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone NOT NULL DEFAULT now();

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS created_at timestamp with time zone NOT NULL DEFAULT now();

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS is_tester boolean;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS facebook_id text;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS platform text;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS phone_verified boolean;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS email_verified boolean;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS given_name TEXT;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS last_login_date timestamp with time zone;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS birthday DATE NULL;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS gender TEXT;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '';

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '';

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS first_name_phonetic TEXT;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS last_name_phonetic TEXT;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS full_name_phonetic TEXT;


-- Drop the default values

ALTER TABLE ONLY public.users ALTER COLUMN country DROP DEFAULT;
ALTER TABLE ONLY public.users ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE ONLY public.users ALTER COLUMN updated_at DROP DEFAULT;