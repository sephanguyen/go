CREATE TABLE IF NOT EXISTS  public.e2e_features_status (
    name text,
    status text[] default '{}',
    count bigint,
    feature_ids text[] default '{}',
    instance_ids text[] default '{}'
);

DROP FUNCTION IF EXISTS public.get_e2e_feature_status_count_in_last_n_days;

CREATE OR REPLACE FUNCTION public.get_e2e_feature_status_count_in_last_n_days(
  _nDay integer = null,
  _tags text = 20,
 _on_trunk boolean = true
) RETURNS SETOF e2e_features_status
    LANGUAGE sql STABLE 
    AS $$
    SELECT 
       distinct on (name) name, 
       (array_agg((eef.status) order by eef.created_at desc))[0:_nDay] as status, 
       count(*) as count, 
       (array_agg((eef.feature_id) order by eef.created_at desc))[0:_nDay] as feature_ids, 
       (array_agg((eef.instance_id) order by eef.created_at desc))[0:_nDay] as instance_ids
    FROM
      e2e_features eef
    WHERE
      CASE
        WHEN _tags is not null
        THEN _tags = ANY((eef.tags) :: text[])
        ELSE 1 = 1
      END AND
      CASE
        WHEN (_on_trunk is true)
        THEN eef.on_trunk = true
        ELSE 1 = 1
      END 
    group by eef.name, eef.tags 
    order by eef.name asc;
    $$;
    