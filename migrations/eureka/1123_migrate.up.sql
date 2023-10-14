CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS learning_material_name_idx_gin_trgm ON public.learning_material USING GIN (name gin_trgm_ops);