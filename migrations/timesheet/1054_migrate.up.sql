ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users__facebook_id__key;
ALTER TABLE IF EXISTS users ADD CONSTRAINT users__facebook_id__key UNIQUE (facebook_id, resource_path);

DROP INDEX IF EXISTS users_full_name_phonetic_idx;
DROP INDEX IF EXISTS users_given_name;
DROP INDEX IF EXISTS users_name_idx;
DROP INDEX IF EXISTS users__created_at__idx_asc_nulls_last;
DROP INDEX IF EXISTS users__created_at__idx_desc_nulls_first;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS users_full_name_phonetic_idx ON public.users USING GIN (full_name_phonetic gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users__name__idx ON users USING GIN (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users__created_at__idx_desc ON users (created_at desc);