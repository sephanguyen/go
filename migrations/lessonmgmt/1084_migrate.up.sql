CREATE INDEX IF NOT EXISTS user_basic_info_full_name_phonetic_idx ON public.user_basic_info USING gin (full_name_phonetic gin_trgm_ops);
CREATE INDEX IF NOT EXISTS user_basic_info__created_at__idx_desc ON public.user_basic_info USING btree (created_at DESC);
CREATE INDEX IF NOT EXISTS user_basic_info__created_at_desc__user_id_desc__idx ON public.user_basic_info USING btree (created_at DESC, user_id DESC);
CREATE INDEX IF NOT EXISTS user_basic_info__email_gin__idx ON public.user_basic_info USING gin (email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS user_basic_info__lower_email__idx ON public.user_basic_info USING btree (lower (email));
CREATE INDEX IF NOT EXISTS user_basic_info_resource_path_idx ON public.user_basic_info USING btree (resource_path);
