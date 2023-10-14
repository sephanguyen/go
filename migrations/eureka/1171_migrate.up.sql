CREATE
OR REPLACE FUNCTION private_search_name_lm_fn(search_name text) RETURNS SETOF public.learning_material AS $$
    SELECT
        *
    FROM
        public.learning_material
    WHERE
        (name ilike search_name)
$$ LANGUAGE SQL STABLE SECURITY DEFINER;

CREATE
OR REPLACE FUNCTION public.filter_rls_search_name_lm_fn(search_name text) RETURNS SETOF public.learning_material AS $$
    SELECT
        sl.*
    FROM
        private_search_name_lm_fn(search_name) AS sl
        JOIN public.learning_material ON sl.learning_material_id = learning_material.learning_material_id 
$$ LANGUAGE SQL STABLE;
