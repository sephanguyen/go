CREATE FUNCTION public.bypass_rls_search_name_user_fn() RETURNS SETOF public.users
    LANGUAGE sql STABLE SECURITY DEFINER
    AS $$
    SELECT
        *
    FROM
        public.users
$$;
