ALTER TABLE ONLY public.granted_role
  DROP CONSTRAINT IF EXISTS pk__granted_role;
ALTER TABLE ONLY public.granted_role
  ADD CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id);

ALTER TABLE public.permission ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE public.permission_role ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
