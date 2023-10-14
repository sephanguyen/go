DROP FUNCTION IF EXISTS get_locations_lowest_level;

CREATE OR REPLACE FUNCTION public.get_locations_lowest_level(location_limit int, location_offset int, location_search_name text) 
returns setof public.locations 
    language sql stable
    as $$
	   	select l.*
			from locations l
		where l."name" ILIKE ('%' || location_search_name || '%') 
			and l.deleted_at is null
			and l.is_archived = false
			and l.location_type not in (select lt.parent_location_type_id from location_types lt
				where lt.parent_location_type_id is not null and lt.deleted_at is null and lt.is_archived = false
			)
		order by l.created_at desc limit location_limit offset location_offset;
$$;