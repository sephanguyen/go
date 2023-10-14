DROP FUNCTION IF EXISTS public.find_schools_based_on_student_in_granted_locations;

CREATE
OR REPLACE FUNCTION find_schools_based_on_student_in_granted_locations(
    logged_in_user_id TEXT,
    keyword TEXT,
    "limit" int,
    "offset" int
) RETURNS SETOF school_info LANGUAGE SQL STABLE 
AS $FUNCTION$
SELECT DISTINCT si.*
FROM school_info si 
    JOIN school_history sh ON si.school_id = sh.school_id AND si.resource_path = sh.resource_path
    JOIN students s ON sh.student_id = s.student_id AND sh.resource_path = s.resource_path
    JOIN user_access_paths uap ON s.student_id = uap.user_id AND s.resource_path = uap.resource_path
WHERE sh.is_current = TRUE
    AND si.school_name ILIKE concat('%', $2, '%')
    AND uap.location_id = ANY (
        SELECT location_id
        FROM granted_permissions gp
        WHERE gp.user_id = $1 AND gp.permission_name = 'user.user.read'
    )
    AND si.deleted_at IS NULL
    AND sh.deleted_at IS NULL
    AND s.deleted_at IS NULL
    AND uap.deleted_at IS NULL
    AND (si.is_archived IS NULL OR si.is_archived IS FALSE)
ORDER BY si.school_name ASC
LIMIT $3 OFFSET $4 
$FUNCTION$
;
