CREATE INDEX IF NOT EXISTS users_name_gist_trgm_idx ON public.users USING gist (name public.gist_trgm_ops);

CREATE OR REPLACE FUNCTION private_search_name_user_fn(search_name text) RETURNS SETOF public.users AS $$
    SELECT
        *
    FROM
        public.users
    WHERE
        (search_name IS NULL OR name ilike '%' || search_name || '%')
$$ LANGUAGE SQL STABLE SECURITY DEFINER;

CREATE OR REPLACE FUNCTION public.filter_rls_search_name_user_fn(search_name text) RETURNS SETOF public.users AS $$
    SELECT
        us.*
    FROM
        private_search_name_user_fn(search_name) AS us
    JOIN public.users ON 
        us.user_id = users.user_id 
$$ LANGUAGE SQL STABLE;

CREATE OR REPLACE FUNCTION private_search_name_exam_lo_fn(search_name text) RETURNS SETOF public.exam_lo AS $$
    SELECT
        *
    FROM
        public.exam_lo
    WHERE
        (search_name IS NULL OR name ilike '%' || search_name || '%')
$$ LANGUAGE SQL STABLE SECURITY DEFINER;

CREATE OR REPLACE FUNCTION public.filter_rls_search_name_exam_lo_fn(search_name text) RETURNS SETOF public.exam_lo AS $$
    SELECT
        el.*
    FROM
        private_search_name_exam_lo_fn(search_name) AS el
    JOIN public.exam_lo ON 
        el.learning_material_id = exam_lo.learning_material_id 
$$ LANGUAGE SQL STABLE;
