CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS courses__name__idx_gin_trgm ON public.courses USING GIN (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS courses__created_at__idx_asc_nulls_last ON public.courses (created_at ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS courses__created_at__idx_desc_nulls_first ON public.courses (created_at DESC NULLS FIRST);
