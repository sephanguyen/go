WITH RECURSIVE type_dept (id, dept) AS (
  SELECT location_type_id, 0
  FROM location_types
  WHERE parent_location_type_id  IS null
  UNION ALL
  SELECT lt.location_type_id , d.dept + 1
  FROM location_types lt
  JOIN type_dept d ON lt.parent_location_type_id  = d.id
)
UPDATE location_types 
SET "level" = type_dept.dept
FROM type_dept
WHERE type_dept.id = location_type_id 