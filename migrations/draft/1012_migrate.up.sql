DROP FUNCTION IF EXISTS public.get_instances_with_filters;

CREATE OR REPLACE FUNCTION public.get_instances_with_filters(
  _status text = null,
  _squad_tags text[] = null,
  _feature_tag text[] = null,
  _environment text = null,
  _date_from timestamptz = null,
  _date_till timestamptz = null
) RETURNS SETOF e2e_instances 
    LANGUAGE sql STABLE 
    AS $$
    SELECT
      eei.*
    FROM
      e2e_instances eei
    WHERE
      CASE 
        WHEN _status is not null
        THEN eei.status = _status
        ELSE 1 = 1
      END AND
      CASE
        WHEN _squad_tags is not null
        THEN _squad_tags <@ ((eei.squad_tags) :: text[])
        ELSE 1 = 1
      END AND
      CASE
        WHEN _feature_tag is not null
        THEN _feature_tag <@ ((eei.tags) :: text[])
        ELSE 1 = 1 
      END AND 
      CASE 
        WHEN _environment is not null
        THEN eei.flavor->> 'env' = _environment
        ELSE 1 = 1 
      END AND 
      CASE
        WHEN ( _date_from is not null AND _date_till is not null )
        THEN eei.created_at BETWEEN _date_from AND _date_till
        ELSE 1 = 1 
      END
    $$;


CREATE TABLE IF NOT EXISTS  public.e2e_instances_status_count (
    status text,
    instances_count bigint
);

DROP FUNCTION IF EXISTS public.count_instances_group_by_status;

CREATE OR REPLACE FUNCTION public.count_instances_group_by_status(
  _squad_tags text[] = null,
  _feature_tag text[] = null,
  _environment text = null,
  _date_from timestamptz = null,
  _date_till timestamptz = null
) RETURNS SETOF e2e_instances_status_count
    LANGUAGE sql STABLE 
    AS $$
    SELECT
      COALESCE(eei.status, 'All Status') AS status,
      COUNT(eei.instance_id) AS "instances_count"
    FROM
      e2e_instances eei
    WHERE
      CASE
        WHEN _squad_tags is not null
        THEN _squad_tags <@ ((eei.squad_tags) :: text[])
        ELSE 1 = 1
      END AND
      CASE
        WHEN _feature_tag is not null
        THEN _feature_tag <@ ((eei.tags) :: text[])
        ELSE 1 = 1
      END AND 
      CASE 
        WHEN _environment is not null
        THEN eei.flavor->> 'env' = _environment
        ELSE 1 = 1
      END AND 
      CASE
        WHEN (_date_from is not null AND _date_till is not null)
        THEN eei.created_at BETWEEN _date_from AND _date_till
        ELSE 1 = 1
      END
    GROUP BY 
      ROLLUP(eei.status)
    $$;
    