ALTER TABLE granted_permission DROP CONSTRAINT IF EXISTS granted_permission__uniq;
ALTER TABLE granted_permission ADD CONSTRAINT granted_permission__pk PRIMARY KEY (user_group_id, role_id, permission_id, location_id);
