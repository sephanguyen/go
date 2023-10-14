CREATE
OR REPLACE FUNCTION private_search_name_lm_fn(search_name text) RETURNS SETOF public.learning_material AS $$
SELECT
    *
FROM
    public.learning_material
WHERE
    name ilike search_name OR search_name IS NULL 
$$ LANGUAGE SQL STABLE SECURITY DEFINER;

