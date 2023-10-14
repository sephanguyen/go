INSERT INTO public.permission_role
(permission_id, role_id, created_at, updated_at, resource_path) (
	SELECT p.permission_id, role_id , NOW() AS created_at , NOW() AS updated_at , r.resource_path
	FROM "permission" p, "role" r
	WHERE r.resource_path = p.resource_path
	AND r.role_name = 'OpenAPI'
	AND r.deleted_at IS NULL
	AND p.deleted_at IS NULL
	AND p.permission_name = ANY('{
		user.user.read,
		user.user.write,
		user.student.read,
		user.student.write,
		user.usergroup.read,
		user.usergroupmember.write,
		master.location.read
	}')
	ORDER BY resource_path
) ON CONFLICT DO NOTHING;
