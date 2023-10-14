CREATE OR REPLACE FUNCTION public.filter_rls_user_fn(search_name text, student_ids text[]) RETURNS SETOF public.users AS $$
    SELECT
        us.*
    FROM
        private_search_name_user_fn(search_name) us
    JOIN users USING(user_id)
    WHERE (student_ids is null or user_id = any(student_ids))
$$ LANGUAGE SQL STABLE;
