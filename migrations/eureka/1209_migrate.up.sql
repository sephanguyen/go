DROP INDEX IF EXISTS users_name_gist_trgm_idx;
DROP INDEX IF EXISTS exam_lo_name_gist_trgm_idx;

CREATE INDEX IF NOT EXISTS exam_lo_name_gin_trgm_idx ON exam_lo USING GIN (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS user_name_gin_trgm_idx ON users USING GIN (name gin_trgm_ops);

CREATE OR REPLACE FUNCTION private_search_name_user_fn(search_name text) RETURNS SETOF public.users AS $$
    SELECT
        *
    FROM
        public.users
    WHERE
        name ilike '%' || search_name || '%'
$$ LANGUAGE SQL STABLE SECURITY DEFINER;

CREATE OR REPLACE FUNCTION public.filter_rls_search_name_user_fn(search_name text) RETURNS SETOF public.users AS $$
    SELECT
        us.*
    FROM
        private_search_name_user_fn(search_name) us
    JOIN users USING(user_id)
$$ LANGUAGE SQL STABLE;

CREATE OR REPLACE FUNCTION private_search_name_exam_lo_fn(search_name text) RETURNS SETOF public.exam_lo AS $$
    SELECT
        *
    FROM
        public.exam_lo
    WHERE
        name ilike '%' || search_name || '%'
$$ LANGUAGE SQL STABLE SECURITY DEFINER;

CREATE OR REPLACE FUNCTION public.filter_rls_search_name_exam_lo_fn(search_name text) RETURNS SETOF public.exam_lo AS $$
    SELECT
        el.*
    FROM
        private_search_name_exam_lo_fn(search_name) el
    JOIN public.exam_lo USING(learning_material_id)
$$ LANGUAGE SQL STABLE;
