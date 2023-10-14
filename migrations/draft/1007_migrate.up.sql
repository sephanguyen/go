CREATE OR REPLACE FUNCTION public.get_instance_filter_by_tags(_run_id text, _squad_tags text[], _feature_tag text[]) RETURNS SETOF e2e_instances
    LANGUAGE sql STABLE
    AS $$
    SELECT eei.* FROM e2e_instances eei 
    WHERE 
    CASE WHEN (_squad_tags) is not null then _squad_tags <@ ((eei.squad_tags)::text[]) end
    and case when (_feature_tag) is not null THEN _feature_tag <@ ((eei.tags)::text[]) end
    and flavor->>'run_id' ilike ('%' || _run_id || '%')
$$;
