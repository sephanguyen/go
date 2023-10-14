-- replicated users table indexes
CREATE INDEX IF NOT EXISTS users__created_at__idx_desc ON public.users (created_at desc);
CREATE INDEX IF NOT EXISTS users__lower_email__idx ON public.users (LOWER(email));
CREATE INDEX IF NOT EXISTS users_name_idx ON public.users (name);
CREATE INDEX IF NOT EXISTS users_given_name ON public.users (given_name);
CREATE INDEX IF NOT EXISTS users_resource_path_idx on users using btree(resource_path);

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS users_full_name_phonetic_idx ON public.users USING GIN (full_name_phonetic gin_trgm_ops);


-- replicated students table indexes
CREATE INDEX IF NOT EXISTS students__created_at__idx_desc ON public.students (created_at desc);
CREATE INDEX IF NOT EXISTS students_resource_path_idx on students using btree(resource_path);


-- replicated user_group_member table index
CREATE INDEX IF NOT EXISTS user_group_user_id_idx ON user_group_member (user_group_id, user_id);


-- replicated user_access_paths table indexes
CREATE INDEX IF NOT EXISTS user_access_paths__location_id__idx ON public.user_access_paths(location_id);
CREATE INDEX IF NOT EXISTS user_access_paths__user_id__idx ON public.user_access_paths (user_id);


-- replicated permission table index
CREATE INDEX IF NOT EXISTS idx__permission__permssion_name on "permission" (permission_name);


-- replicated ranted_role table index
CREATE INDEX IF NOT EXISTS granted_role_user_group_id_idx ON public.granted_role USING btree (user_group_id);


-- replicated granted_permission table index
CREATE INDEX IF NOT EXISTS granted_permission__user_group_id__idx ON public.granted_permission USING btree (user_group_id);
CREATE INDEX IF NOT EXISTS granted_permission__role_name__idx ON public.granted_permission USING btree (role_name);
CREATE INDEX IF NOT EXISTS granted_permission__permission_name__idx ON public.granted_permission USING btree (permission_name);


-- invoice table index
CREATE INDEX IF NOT EXISTS invoice__created_at__idx_desc ON public.invoice (created_at desc);
CREATE INDEX IF NOT EXISTS invoice__status__idx ON public.invoice (status);
CREATE INDEX IF NOT EXISTS invoice__type__idx ON public.invoice (type);
CREATE INDEX IF NOT EXISTS invoice__student_id__idx on public.invoice (student_id); 


-- payment table index
CREATE INDEX IF NOT EXISTS payment__invoice_id__idx ON public.payment (invoice_id);
CREATE INDEX IF NOT EXISTS payment__payment_method__idx ON public.payment (payment_method);