ALTER TABLE public.granted_role ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE public.granted_role_access_path ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE public.role ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE public.user_group ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
