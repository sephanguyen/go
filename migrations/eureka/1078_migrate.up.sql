CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS learning_objectives_name_idx_gin_trgm ON public.learning_objectives USING GIN (name gin_trgm_ops);
