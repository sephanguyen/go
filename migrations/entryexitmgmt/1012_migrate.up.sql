CREATE TABLE IF NOT EXISTS public.role (
    role_id TEXT NOT NULL,
    role_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__role PRIMARY KEY (role_id)
);
CREATE POLICY rls_role ON public.role USING (permission_check(resource_path, 'role')) WITH CHECK (permission_check(resource_path, 'role'));
ALTER TABLE public.role ENABLE ROW LEVEL security;
ALTER TABLE public.role FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.user_group (
    user_group_id TEXT NOT NULL,
    user_group_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__user_group PRIMARY KEY (user_group_id)
);
CREATE POLICY rls_user_group ON public.user_group USING (permission_check(resource_path, 'user_group')) WITH CHECK (permission_check(resource_path, 'user_group'));
ALTER TABLE public.user_group ENABLE ROW LEVEL security;
ALTER TABLE public.user_group FORCE ROW LEVEL security;

ALTER TABLE public.user_group ADD COLUMN IF NOT EXISTS org_location_id TEXT;
ALTER TABLE public.user_group ADD CONSTRAINT fk__user_group__org_location_id FOREIGN KEY (org_location_id) REFERENCES public.locations(location_id);


CREATE TABLE IF NOT EXISTS public.granted_role (
    granted_role_id TEXT NOT NULL UNIQUE,
    user_group_id TEXT NOT NULL,
    role_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__granted_role PRIMARY KEY (user_group_id, role_id),
    CONSTRAINT fk__granted_role__user_group_id FOREIGN KEY (user_group_id) REFERENCES public.user_group(user_group_id),
    CONSTRAINT fk__granted_role__role_id FOREIGN KEY (role_id) REFERENCES public.role(role_id)
);
CREATE POLICY rls_granted_role ON public.granted_role USING (permission_check(resource_path, 'granted_role')) WITH CHECK (permission_check(resource_path, 'granted_role'));
ALTER TABLE public.granted_role ENABLE ROW LEVEL security;
ALTER TABLE public.granted_role FORCE ROW LEVEL security;

ALTER TABLE ONLY public.granted_role
  DROP CONSTRAINT IF EXISTS pk__granted_role;
ALTER TABLE ONLY public.granted_role
  ADD CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id);



CREATE TABLE IF NOT EXISTS public.user_group_member (
    user_id TEXT NOT NULL,
    user_group_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id),
    CONSTRAINT fk__user_group_member__user_id FOREIGN KEY (user_id) REFERENCES public.users(user_id),
    CONSTRAINT fk__user_group_member__user_group_id FOREIGN KEY (user_group_id) REFERENCES public.user_group(user_group_id)
);
CREATE POLICY rls_user_group_member ON public.user_group_member USING (permission_check(resource_path, 'user_group_member')) with check (permission_check(resource_path, 'user_group_member'));
ALTER TABLE public.user_group_member ENABLE ROW LEVEL security;
ALTER TABLE public.user_group_member FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.permission (
    permission_id TEXT NOT NULL,
    permission_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__permission PRIMARY KEY (permission_id)
);
CREATE POLICY rls_permission ON public.permission USING (permission_check(resource_path, 'permission')) WITH CHECK (permission_check(resource_path, 'permission'));
ALTER TABLE public.permission ENABLE ROW LEVEL security;
ALTER TABLE public.permission FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.permission_role (
    permission_id TEXT NOT NULL,
    role_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__permission_role PRIMARY KEY (permission_id, role_id),
    CONSTRAINT fk__permission_role__permission_id FOREIGN KEY (permission_id) REFERENCES public.permission(permission_id),
    CONSTRAINT fk__permission_role__role_id FOREIGN KEY (role_id) REFERENCES public.role(role_id)
);
CREATE POLICY rls_permission_role ON public.permission_role USING (permission_check(resource_path, 'permission_role')) WITH CHECK (permission_check(resource_path, 'permission_role'));
ALTER TABLE public.permission_role ENABLE ROW LEVEL security;
ALTER TABLE public.permission_role FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.granted_role_access_path (
    granted_role_id TEXT NOT NULL,
    location_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__granted_role_access_path PRIMARY KEY (granted_role_id, location_id),
    CONSTRAINT fk__granted_role_access_path__location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id),
    CONSTRAINT fk__granted_role_access_path__granted_role_id FOREIGN KEY (granted_role_id) REFERENCES public.granted_role(granted_role_id)
);
CREATE POLICY rls_granted_role_access_path ON public.granted_role_access_path USING (permission_check(resource_path, 'granted_role_access_path')) WITH CHECK (permission_check(resource_path, 'granted_role_access_path'));
ALTER TABLE public.granted_role_access_path ENABLE ROW LEVEL security;
ALTER TABLE public.granted_role_access_path FORCE ROW LEVEL security; 