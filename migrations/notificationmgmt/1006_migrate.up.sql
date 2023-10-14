CREATE
OR REPLACE VIEW public.granted_permissions AS
SELECT
    ugm.user_id,
    p.permission_name,
    l1.location_id,
    ugm.resource_path,
    p.permission_id
FROM
    user_group_member ugm
    JOIN user_group ug ON ugm.user_group_id = ug.user_group_id
    JOIN granted_role gr ON ug.user_group_id = gr.user_group_id
    JOIN ROLE r ON gr.role_id = r.role_id
    JOIN permission_role pr ON r.role_id = pr.role_id
    JOIN PERMISSION p ON p.permission_id = pr.permission_id
    JOIN granted_role_access_path grap ON gr.granted_role_id = grap.granted_role_id
    JOIN locations l ON l.location_id = grap.location_id
    JOIN locations l1 ON l1.access_path ~~ (l.access_path || '%' :: TEXT)
WHERE
    ugm.deleted_at IS NULL
    AND ug.deleted_at IS NULL
    AND gr.deleted_at IS NULL
    AND r.deleted_at IS NULL
    AND pr.deleted_at IS NULL
    AND p.deleted_at IS NULL
    AND grap.deleted_at IS NULL
    AND l.deleted_at IS NULL
    AND l1.deleted_at IS NULL;