DROP FUNCTION IF EXISTS get_lowest_location_types;

CREATE
OR REPLACE FUNCTION public.get_lowest_location_types() returns setof public.location_types language sql stable as $$
SELECT
  *
from
  location_types
WHERE
  location_type_id not in (
    SELECT
      distinct parent_location_type_id
    from
      location_types
    WHERE
      parent_location_type_id is not null
      AND deleted_at is null
  )
  AND deleted_at is null;
$$;
