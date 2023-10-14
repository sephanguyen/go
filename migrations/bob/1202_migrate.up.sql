insert
	into
	class_member (class_member_id,
	class_id,
	user_id,
    created_at,
    updated_at,
	resource_path)
select
    distinct on (cm.class_member_id::text)
	cm.class_member_id::text,
	cm.class_id ::text,
    cm.user_id::text,
    cm.created_at,
    cm.updated_at,
    cm.resource_path
from
	class_members cm
join class c on
	c.class_id = cm.class_id::text
where
	c.deleted_at is null
	and cm.deleted_at is null
	and cm.status = 'CLASS_MEMBER_STATUS_ACTIVE'
    and c.resource_path is not null
    and cm.resource_path is not null
    and cm.deleted_at is null
ON CONFLICT DO NOTHING;
