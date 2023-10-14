DROP INDEX IF EXISTS user_group_member_user_group_id_idx;
CREATE INDEX IF NOT EXISTS user_group_member_user_group_id_idx ON public.user_group_member USING btree (user_group_id);
