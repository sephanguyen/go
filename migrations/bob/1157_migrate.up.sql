ALTER TABLE locations ADD COLUMN IF NOT EXISTS access_path TEXT;

-- update access_path old locations --

WITH RECURSIVE with_locations(location_id, name, parent_location_id, access_path) AS (
    SELECT l.location_id , l."name" , l.parent_location_id , l.location_id::TEXT AS access_path 
    FROM locations AS l 
    WHERE l.parent_location_id IS NULL
UNION ALL
    SELECT lo.location_id, lo."name", lo.parent_location_id, (wl.access_path || '/' || lo.location_id::TEXT) 
    FROM with_locations AS wl, locations AS lo 
    WHERE lo.parent_location_id = wl.location_id
) UPDATE locations 
set access_path = with_locations.access_path
from with_locations
where with_locations.location_id = locations.location_id;
