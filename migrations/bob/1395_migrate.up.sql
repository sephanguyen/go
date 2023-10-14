INSERT INTO permission_role(
 role_id, permission_id, created_at, updated_at, deleted_at, resource_path
) (
SELECT
  role_id, permission_id, now() created_at, now() updated_at, null deleted_at, role.resource_path resource_path
FROM role, permission
WHERE
  role_name = ANY(ARRAY['Teacher Lead', 'Teacher', 'Student', 'Parent']) AND
  permission_name = 'user.usergroup.read' AND
  role.resource_path = permission.resource_path AND
  permission.deleted_at IS NULL AND
  role.deleted_at IS NULL
) ON CONFLICT DO NOTHING
