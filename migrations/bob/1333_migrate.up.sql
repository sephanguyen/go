-- insert data for granted_permission table
INSERT INTO public.granted_permission(
	user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT ug.user_group_id, ug.user_group_name, r.role_id, r.role_name, p.permission_id, p.permission_name, grap.location_id, ug.resource_path 
FROM user_group ug
LEFT JOIN granted_role gr 
  ON ug.user_group_id = gr.user_group_id AND gr.deleted_at IS NULL
LEFT JOIN granted_role_access_path grap 
  ON gr.granted_role_id = grap.granted_role_id AND grap.deleted_at IS NULL
LEFT JOIN role r 
  ON gr.role_id = r.role_id AND r.deleted_at IS NULL
LEFT JOIN permission_role pr 
  ON r.role_id = pr.role_id AND pr.deleted_at IS NULL
LEFT JOIN permission p 
  ON pr.permission_id = p.permission_id AND p.deleted_at IS NULL
WHERE ug.deleted_at IS NULL
  AND gr.granted_role_id IS NOT NULL 
  AND grap.granted_role_id IS NOT NULL
  AND r.role_id IS NOT NULL
  AND pr.permission_id IS NOT NULL
  AND p.permission_id IS NOT NULL
  AND ug.resource_path != '-2147483644'
ON CONFLICT ON CONSTRAINT granted_permission__uniq DO NOTHING;
