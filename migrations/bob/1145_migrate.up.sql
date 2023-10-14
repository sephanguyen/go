CREATE INDEX IF NOT EXISTS users__created_at__idx_asc_nulls_last ON public.users (created_at ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS users__created_at__idx_desc_nulls_first ON public.users (created_at DESC NULLS FIRST);
