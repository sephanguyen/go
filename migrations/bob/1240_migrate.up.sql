DROP INDEX IF EXISTS staff__created_at__idx_desc;
CREATE INDEX IF NOT EXISTS staff__created_at__idx_desc ON public.staff (created_at desc);
