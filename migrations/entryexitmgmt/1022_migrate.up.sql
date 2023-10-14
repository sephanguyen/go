ALTER TABLE public.role DROP CONSTRAINT IF EXISTS pk__role;

ALTER TABLE public.role ADD CONSTRAINT pk__role PRIMARY KEY (role_id, resource_path);

ALTER TABLE public.permission_role DROP CONSTRAINT IF EXISTS pk__permission_role;

ALTER TABLE public.permission_role ADD CONSTRAINT pk__permission_role PRIMARY KEY (permission_id, role_id, resource_path);