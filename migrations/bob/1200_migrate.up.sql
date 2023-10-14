insert
	into
	class (class_id,
	"name",
	location_id,
	course_id,
	school_id ,
    created_at,
    updated_at,
	resource_path)
select
    distinct on (c.class_id::text)
	c.class_id::text,
	c."name",
	'01FR4M51XJY9E77GSN4QZ1Q9N2',
	cc.course_id ,
	c.school_id::text,
    c.created_at,
    c.updated_at,
	c.resource_path
from
	classes c
join courses_classes cc on
	cc.class_id = c.class_id
join courses co on
	co.course_id = cc.course_id 
where
	c.deleted_at is null
	and cc.deleted_at is null
	and c.status = 'CLASS_STATUS_ACTIVE'
    and c.resource_path is not null
    and c.deleted_at is null
ON CONFLICT DO NOTHING;