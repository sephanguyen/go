-- users
DROP INDEX IF EXISTS users__email__resource_path__idx;
DROP INDEX IF EXISTS users_given_name;
DROP INDEX IF EXISTS resource_path_idx;

DROP INDEX IF EXISTS users_name_idx;
CREATE INDEX IF NOT EXISTS users__name__idx ON public.users USING GIN (name gin_trgm_ops);

-- students
DROP INDEX IF EXISTS students__created_at__btree;
DROP INDEX IF EXISTS students__created_at__idx_asc;

-- rename staff index
DROP INDEX IF EXISTS staff__staff_id__idx;
CREATE INDEX IF NOT EXISTS staff__resource_path__idx on staff using btree(resource_path);
