CREATE TABLE public.users (
	user_id text NOT NULL,
	"name" text NOT NULL,
	user_group text NOT NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	country text NOT NULL,
	avatar text NULL,
	phone_number text NULL,
	email text NULL,
	device_token text NULL,
	allow_notification bool NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	is_tester bool NULL,
	facebook_id text NULL,
	platform text NULL,
	phone_verified bool NULL,
	email_verified bool NULL,
	deleted_at TIMESTAMP WITH TIME ZONE NULL,
	given_name text NULL,
	last_login_date TIMESTAMP WITH TIME ZONE NULL,
	birthday date NULL,
	gender text NULL,
	first_name text NOT NULL DEFAULT '',
	last_name text NOT NULL DEFAULT '',
	first_name_phonetic text NULL,
	last_name_phonetic text NULL,
	full_name_phonetic text NULL,
	CONSTRAINT users_pk PRIMARY KEY (user_id)
);
CREATE POLICY rls_users ON public.users USING (permission_check(resource_path, 'users')) WITH CHECK (permission_check(resource_path, 'users'));
ALTER TABLE public.users ENABLE ROW LEVEL security;
ALTER TABLE public.users FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.role (
    role_id TEXT NOT NULL,
    role_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

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
    org_location_id TEXT,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__user_group PRIMARY KEY (user_group_id),
    CONSTRAINT fk__user_group__org_location_id FOREIGN KEY (org_location_id) REFERENCES public.locations(location_id)
);
CREATE POLICY rls_user_group ON public.user_group USING (permission_check(resource_path, 'user_group')) WITH CHECK (permission_check(resource_path, 'user_group'));
ALTER TABLE public.user_group ENABLE ROW LEVEL security;
ALTER TABLE public.user_group FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.granted_role (
    granted_role_id TEXT NOT NULL UNIQUE,
    user_group_id TEXT NOT NULL,
    role_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),
    
    CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id),
    CONSTRAINT fk__granted_role__user_group_id FOREIGN KEY (user_group_id) REFERENCES public.user_group(user_group_id),
    CONSTRAINT fk__granted_role__role_id FOREIGN KEY (role_id) REFERENCES public.role(role_id)
);
CREATE POLICY rls_granted_role ON public.granted_role USING (permission_check(resource_path, 'granted_role')) WITH CHECK (permission_check(resource_path, 'granted_role'));
ALTER TABLE public.granted_role ENABLE ROW LEVEL security;
ALTER TABLE public.granted_role FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.user_group_member (
    user_id TEXT NOT NULL,
    user_group_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id),
    CONSTRAINT fk__user_group_member__user_id FOREIGN KEY (user_id) REFERENCES public.users(user_id),
    CONSTRAINT fk__user_group_member__user_group_id FOREIGN KEY (user_group_id) REFERENCES public.user_group(user_group_id)
);
CREATE POLICY rls_user_group_member ON public.user_group_member USING (permission_check(resource_path, 'user_group_member')) with check (permission_check(resource_path, 'user_group_member'));
ALTER TABLE public.user_group_member ENABLE ROW LEVEL security;
ALTER TABLE public.user_group_member FORCE ROW LEVEL security;
