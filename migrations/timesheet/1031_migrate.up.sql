ALTER TABLE public.role DROP CONSTRAINT IF EXISTS pk__role;

ALTER TABLE public.role ADD CONSTRAINT role__pk PRIMARY KEY (role_id, resource_path);