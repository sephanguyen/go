DROP INDEX IF EXISTS staff__created_at__idx_desc;
CREATE INDEX IF NOT EXISTS staff__created_at__idx_desc ON public.staff (created_at desc);

DROP INDEX IF EXISTS students__created_at__idx_desc;
CREATE INDEX IF NOT EXISTS students__created_at__idx_desc ON public.students (created_at desc);

DROP INDEX IF EXISTS users__created_at__idx_desc;
CREATE INDEX IF NOT EXISTS users__created_at__idx_desc ON public.users (created_at desc);

DROP INDEX IF EXISTS user_access_paths__user_id__idx;
CREATE INDEX IF NOT EXISTS user_access_paths__user_id__idx ON public.user_access_paths (user_id);

DROP INDEX IF EXISTS staff__staff_id__idx;
CREATE INDEX IF NOT EXISTS staff__staff_id__idx on staff using btree(resource_path);
