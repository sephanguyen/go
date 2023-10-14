CREATE
OR REPLACE FUNCTION public.get_lowest_location_types_v2() returns setof public.location_types language sql stable as $$
SELECT *
FROM location_types l1
where l1.deleted_at is null  and "level" = (SELECT MAX("level") FROM location_types l2 WHERE l2.resource_path  = l1.resource_path and l2.deleted_at is null)
GROUP BY resource_path, location_type_id
$$;
