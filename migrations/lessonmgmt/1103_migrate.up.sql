DROP INDEX IF EXISTS idx_user_id_user_group_member;
DROP INDEX IF EXISTS locations_access_path_text_pattern_ops_idx;
DROP INDEX IF EXISTS permission_name_idx;
CREATE UNIQUE INDEX locations_pkey ON public.locations USING btree (location_id);
