DROP INDEX IF EXISTS users_full_name_phonetic_idx;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS users_full_name_phonetic_idx ON public.users USING GIN (full_name_phonetic gin_trgm_ops);