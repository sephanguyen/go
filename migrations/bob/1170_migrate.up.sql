DROP FUNCTION IF EXISTS get_locations_active_by_course_id;

CREATE OR REPLACE FUNCTION public.get_locations_active_by_course_id(course text) 
returns setof public.course_access_paths 
    language sql stable
    as $$
        SELECT cap.* FROM course_access_paths cap 
        JOIN locations l ON l.location_id = cap.location_id
        WHERE cap.course_id = course AND l.is_archived = false AND
        l.deleted_at is null AND cap.deleted_at is null
;
$$;