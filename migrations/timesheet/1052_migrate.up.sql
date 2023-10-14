CREATE INDEX IF NOT EXISTS granted_role_user_group_id_idx ON granted_role USING btree (user_group_id);
ALTER TABLE granted_role ADD CONSTRAINT granted_role_granted_role_id_key UNIQUE (granted_role_id);
