CREATE INDEX courses__created_at__idx_asc_nulls_last ON public.courses USING btree (created_at);
CREATE INDEX courses__created_at__idx_desc_nulls_first ON public.courses USING btree (created_at DESC);
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS courses__name__idx_gin_trgm ON public.courses USING GIN (name gin_trgm_ops);
CREATE INDEX courses__search_by_name__idx ON public.courses USING btree (created_at DESC, display_order, "left"(name, 256), "left"(course_id, 256));
