CREATE TABLE IF NOT EXISTS public.users (
	user_id text NOT NULL,
	"name" text NOT NULL,
	user_group text NOT NULL,
	avatar text NULL,
	phone_number text NULL,
	email text NULL,
	phone_verified bool NULL,
	email_verified bool NULL,
	given_name text NULL,
	last_login_date TIMESTAMP WITH TIME ZONE NULL,
	birthday date NULL,
	gender text NULL,
	first_name text NOT NULL DEFAULT '',
	last_name text NOT NULL DEFAULT '',
	first_name_phonetic text NULL,
	last_name_phonetic text NULL,
	full_name_phonetic text NULL,
	previous_name text NULL,
	is_tester bool NULL,
	is_system bool NULL,
	user_external_id text NULL,
    deactivated_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),

	CONSTRAINT users_pk PRIMARY KEY (user_id)
);
CREATE POLICY rls_users ON public.users USING (permission_check(resource_path, 'users')) WITH CHECK (permission_check(resource_path, 'users'));
CREATE POLICY rls_users_restrictive ON "users" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'users')) WITH CHECK (permission_check(resource_path, 'users'));
ALTER TABLE public.users ENABLE ROW LEVEL security;
ALTER TABLE public.users FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.role (
    role_id TEXT NOT NULL,
    role_name TEXT NOT NULL,
    is_system bool NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__role PRIMARY KEY (role_id, resource_path)
);
CREATE POLICY rls_role ON public.role USING (permission_check(resource_path, 'role')) WITH CHECK (permission_check(resource_path, 'role'));
CREATE POLICY rls_role_restrictive ON "role" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'role')) WITH CHECK (permission_check(resource_path, 'role'));
ALTER TABLE public.role ENABLE ROW LEVEL security;
ALTER TABLE public.role FORCE ROW LEVEL security;



CREATE TABLE IF NOT EXISTS public.user_group (
    user_group_id TEXT NOT NULL,
    user_group_name TEXT NOT NULL,
    is_system bool NULL,
    org_location_id TEXT,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__user_group PRIMARY KEY (user_group_id)
);
CREATE POLICY rls_user_group ON public.user_group USING (permission_check(resource_path, 'user_group')) WITH CHECK (permission_check(resource_path, 'user_group'));
CREATE POLICY rls_user_group_restrictive ON "user_group" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'user_group')) WITH CHECK (permission_check(resource_path, 'user_group'));
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
    
    CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id)
);
CREATE POLICY rls_granted_role ON public.granted_role USING (permission_check(resource_path, 'granted_role')) WITH CHECK (permission_check(resource_path, 'granted_role'));
CREATE POLICY rls_granted_role_restrictive ON "granted_role" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'granted_role')) WITH CHECK (permission_check(resource_path, 'granted_role'));
ALTER TABLE public.granted_role ENABLE ROW LEVEL security;
ALTER TABLE public.granted_role FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.user_group_member (
    user_id TEXT NOT NULL,
    user_group_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id)
);
CREATE POLICY rls_user_group_member ON public.user_group_member USING (permission_check(resource_path, 'user_group_member')) with check (permission_check(resource_path, 'user_group_member'));
CREATE POLICY rls_user_group_member_restrictive ON "user_group_member" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'user_group_member')) WITH CHECK (permission_check(resource_path, 'user_group_member'));
ALTER TABLE public.user_group_member ENABLE ROW LEVEL security;
ALTER TABLE public.user_group_member FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.granted_role_access_path (
	granted_role_id text NOT NULL,
	location_id text NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),

	CONSTRAINT pk__granted_role_access_path PRIMARY KEY (granted_role_id, location_id)
);
CREATE POLICY rls_granted_role_access_path ON public.granted_role_access_path USING (permission_check(resource_path, 'granted_role_access_path')) with check (permission_check(resource_path, 'granted_role_access_path'));
CREATE POLICY rls_granted_role_access_path_restrictive ON "granted_role_access_path" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'granted_role_access_path')) WITH CHECK (permission_check(resource_path, 'granted_role_access_path'));
ALTER TABLE public.granted_role_access_path ENABLE ROW LEVEL security;
ALTER TABLE public.granted_role_access_path FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public."permission" (
	permission_id text NOT NULL,
	permission_name text NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	
    CONSTRAINT pk__permission PRIMARY KEY (permission_id)
);
CREATE POLICY rls_permission ON public.permission USING (permission_check(resource_path, 'permission')) with check (permission_check(resource_path, 'permission'));
CREATE POLICY rls_permission_restrictive ON "permission" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'permission')) WITH CHECK (permission_check(resource_path, 'permission'));
ALTER TABLE public.permission ENABLE ROW LEVEL security;
ALTER TABLE public.permission FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.permission_role (
	permission_id text NOT NULL,
	role_id text NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	
    CONSTRAINT permission_role__pk PRIMARY KEY (permission_id, role_id, resource_path)
);
CREATE POLICY rls_permission_role ON public.permission_role USING (permission_check(resource_path, 'permission_role')) with check (permission_check(resource_path, 'permission_role'));
CREATE POLICY rls_permission_role_restrictive ON "permission_role" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'permission_role')) WITH CHECK (permission_check(resource_path, 'permission_role'));
ALTER TABLE public.permission_role ENABLE ROW LEVEL security;
ALTER TABLE public.permission_role FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.user_access_paths (
    user_id text NOT NULL,
    location_id text NOT NULL,
    access_path text,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path text DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT user_access_paths_pk PRIMARY KEY (user_id, location_id)
);
CREATE POLICY rls_user_access_paths ON "user_access_paths" using (permission_check(resource_path, 'user_access_paths')) with check (permission_check(resource_path, 'user_access_paths'));
CREATE POLICY rls_user_access_paths_restrictive ON "user_access_paths" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'user_access_paths')) with check (permission_check(resource_path, 'user_access_paths'));
ALTER TABLE "user_access_paths" ENABLE ROW LEVEL security;
ALTER TABLE "user_access_paths" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.locations (
	location_id text NOT NULL,
	"name" text NOT NULL,
	location_type text NULL,
	partner_internal_id text NULL,
	partner_internal_parent_id text NULL,
	parent_location_id text NULL,
	is_archived bool NOT NULL DEFAULT false,
	access_path text NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at TIMESTAMP WITH TIME ZONE NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),

	CONSTRAINT locations_pkey PRIMARY KEY (location_id)
);
CREATE POLICY rls_locations ON "locations" using (permission_check(resource_path, 'locations')) with check (permission_check(resource_path, 'locations'));
CREATE POLICY rls_locations_restrictive ON "locations" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'locations')) with check (permission_check(resource_path, 'locations'));
ALTER TABLE "locations" ENABLE ROW LEVEL security;
ALTER TABLE "locations" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.location_types (
	location_type_id text NOT NULL,
	"name" text NOT NULL,
	display_name text NULL,
	parent_name text NULL,
	parent_location_type_id text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	is_archived bool NOT NULL DEFAULT false,
	"level" int4 NULL DEFAULT 0,

	CONSTRAINT location_types_pkey PRIMARY KEY (location_type_id),
	CONSTRAINT unique__location_type_name_resource_path UNIQUE (name, resource_path),
	CONSTRAINT location_type_id_fk FOREIGN KEY (parent_location_type_id) REFERENCES public.location_types(location_type_id)
);
CREATE POLICY rls_location_types ON "location_types" using (permission_check(resource_path, 'location_types')) with check (permission_check(resource_path, 'location_types'));
CREATE POLICY rls_location_types_restrictive ON "location_types" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'location_types')) with check (permission_check(resource_path, 'location_types'));
ALTER TABLE "location_types" ENABLE ROW LEVEL security;
ALTER TABLE "location_types" FORCE ROW LEVEL security;
