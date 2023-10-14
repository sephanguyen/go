CREATE OR REPLACE FUNCTION public.get_locations_lowest_level_by_name_and_access_path(location_limit integer, location_offset integer, location_search_name text, granted_locations_id_pattern text)
 RETURNS SETOF locations
 LANGUAGE sql
 STABLE
AS $function$
	   	select l.*
			from locations l
		where l."name" ILIKE ('%' || location_search_name || '%') 
            AND l.access_path SIMILAR TO ('%' || granted_locations_id_pattern || '%')
			and l.deleted_at is null
			and l.is_archived = false
			and l.location_type not in (select lt.parent_location_type_id from location_types lt
				where lt.parent_location_type_id is not null and lt.deleted_at is null and lt.is_archived = false
			)
		order by l.created_at desc, l.name asc limit location_limit offset location_offset;
$function$;
