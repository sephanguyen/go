ALTER TABLE public.users ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '';
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '';
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS first_name_phonetic TEXT;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS last_name_phonetic TEXT;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS full_name_phonetic TEXT;

CREATE INDEX users_full_name_phonetic_idx ON public.users USING btree(full_name_phonetic);
