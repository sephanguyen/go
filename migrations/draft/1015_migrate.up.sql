CREATE TABLE IF NOT EXISTS public.e2e_instances_group_by_date (
    created_date text,
    status text,
    instances_count bigint
);

DROP FUNCTION IF EXISTS public.count_instances_group_by_date;

CREATE OR REPLACE FUNCTION public.count_instances_group_by_date (
  _squad_tags text[] = null,
  _feature_tag text[] = null,
  _environment text = null,
  _date_from timestamptz = null,
  _date_till timestamptz = null,
  _group_by text = null
) RETURNS SETOF e2e_instances_group_by_date
    LANGUAGE sql STABLE 
    AS $$
    SELECT
      TO_CHAR(eei.created_at,
              CASE
                WHEN _group_by is not null
                THEN _group_by
                ELSE 'YYYY-MM'
              END
             ) AS created_date,
      eei.status AS status,
      COUNT(eei.instance_id) AS instances_count
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
        created_date, status
    ORDER BY
        created_date DESC
    $$;
    