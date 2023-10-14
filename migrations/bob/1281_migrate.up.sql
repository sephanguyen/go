ALTER TABLE public.permission_role DROP CONSTRAINT IF EXISTS fk__permission_role__role_id;

ALTER TABLE public.granted_role DROP CONSTRAINT IF EXISTS fk__granted_role__role_id;

ALTER TABLE public.role DROP CONSTRAINT IF EXISTS pk__role;

ALTER TABLE public.role ADD CONSTRAINT role__pk PRIMARY KEY (role_id, resource_path);

ALTER TABLE public.permission_role DROP CONSTRAINT IF EXISTS pk__permission_role;

ALTER TABLE public.permission_role ADD CONSTRAINT permission_role__pk PRIMARY KEY (permission_id, role_id, resource_path);

ALTER TABLE public.permission_role 
    ADD CONSTRAINT permission_role__role_id__resource_path__fk 
    FOREIGN KEY (role_id, resource_path) 
    REFERENCES public.role(role_id, resource_path) ON UPDATE CASCADE;

ALTER TABLE public.granted_role 
    ADD CONSTRAINT granted_role__role_id__resource_path__fk 
    FOREIGN KEY (role_id, resource_path) 
    REFERENCES public.role(role_id, resource_path) ON UPDATE CASCADE;
