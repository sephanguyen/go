ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS avatar text;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS phone_number text;

ALTER TABLE ONLY public.users
    ADD COLUMN IF NOT EXISTS email text;

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