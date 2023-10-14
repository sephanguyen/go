ALTER TABLE ONLY public.users DROP CONSTRAINT IF EXISTS user_gender_check;
ALTER TABLE ONLY public.users ADD CONSTRAINT user_gender_check CHECK ((gender = ANY ('{MALE, FEMALE}'::text[])));