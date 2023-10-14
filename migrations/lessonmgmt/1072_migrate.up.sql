CREATE INDEX user_group__user_group_name_idx ON public.user_group USING btree (user_group_name);
CREATE INDEX idx_user_id_user_group_member ON public.user_group_member USING btree (user_id);
CREATE INDEX user_group_user_id_idx ON public.user_group_member USING btree (user_group_id, user_id);
