CREATE TABLE IF NOT EXISTS public.locations
(
    location_id                TEXT                                                          NOT NULL PRIMARY KEY,
    name                       TEXT                                                          NOT NULL,
    CREATED_AT                 TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('UTC'::TEXT, NOW()) NOT NULL,
    updated_at                 TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    deleted_at                 TIMESTAMP WITH TIME ZONE,
    resource_path              TEXT                     DEFAULT autofillresourcepath(),
    location_type              TEXT,
    partner_internal_id        TEXT,
    partner_internal_parent_id TEXT,
    parent_location_id         TEXT,
    IS_ARCHIVED                BOOLEAN                  DEFAULT FALSE                        NOT NULL,
    access_path                TEXT
);
 
-- ALTER TABLE public.locations
--     ADD CONSTRAINT fk__locations_parent_location_id
--     FOREIGN KEY (parent_location_id)
--     REFERENCES public.locations(location_id);

CREATE POLICY rls_locations ON "locations" USING (permission_check(resource_path, 'locations')) WITH CHECK (permission_check(resource_path, 'locations'));
ALTER TABLE IF EXISTS "locations"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS "locations"
    FORCE ROW LEVEL SECURITY;

CREATE TABLE IF NOT EXISTS public.users (
    user_id text NOT NULL,
    name text NOT NULL,
    user_group text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT users_pk PRIMARY KEY (user_id)
);

CREATE POLICY rls_users ON "users" USING (permission_check(resource_path, 'users')) WITH CHECK (permission_check(resource_path, 'users'));

ALTER TABLE IF EXISTS "users" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "users" FORCE ROW LEVEL security;


-- Clone role table from usermgmt
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
ALTER TABLE IF EXISTS public.role ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.role FORCE ROW LEVEL security;

-- Clone user_group table from usermgmt
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
ALTER TABLE IF EXISTS public.user_group ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.user_group FORCE ROW LEVEL security;

ALTER TABLE IF EXISTS public.user_group ADD COLUMN IF NOT EXISTS org_location_id TEXT;
-- ALTER TABLE public.user_group ADD CONSTRAINT fk__user_group__org_location_id FOREIGN KEY (org_location_id) REFERENCES public.locations(location_id);

-- Clone groups table from usermgmt
CREATE TABLE IF NOT EXISTS public.groups (
    group_id text NOT NULL,
    name text NOT NULL,
    description text,
    privileges JSONB,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path TEXT,
    CONSTRAINT pk__groups PRIMARY KEY (group_id)
);
CREATE POLICY rls_groups ON public.groups USING (permission_check(resource_path, 'groups')) WITH CHECK (permission_check(resource_path, 'groups'));
ALTER TABLE IF EXISTS public.groups ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.groups FORCE ROW LEVEL security;

-- Clone users_groups table from usermgmt
CREATE TABLE IF NOT EXISTS public.users_groups (
    user_id text NOT NULL,
    group_id text NOT NULL,
    is_origin bool NOT NULL,
    status TEXT NOT NULL, -- USER_GROUP_STATUS_ACTIVE, USER_GROUP_STATUS_INACTIVE
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path TEXT,
    CONSTRAINT pk__users_groups PRIMARY KEY (user_id, group_id)
    -- CONSTRAINT fk__users_groups__user_id FOREIGN KEY (user_id) REFERENCES public.users (user_id),
    -- CONSTRAINT fk__users_groups__group_id FOREIGN KEY (group_id) REFERENCES public.groups (group_id)
);
CREATE POLICY rls_users_groups ON public.users_groups USING (permission_check(resource_path, 'users_groups')) WITH CHECK (permission_check(resource_path, 'users_groups'));
ALTER TABLE IF EXISTS public.users_groups ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.users_groups FORCE ROW LEVEL security;

-- Clone school_admins table from usermgmt
-- CREATE TABLE IF NOT EXISTS public.school_admins (
--     school_admin_id text NOT NULL,
--     school_id integer NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE NOT NULL,
--     updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
--     deleted_at TIMESTAMP WITH TIME ZONE,
--     resource_path TEXT,
--     CONSTRAINT school_admins_pk PRIMARY KEY (school_admin_id)
-- );
-- CREATE POLICY rls_school_admins ON public.school_admins USING (permission_check(resource_path, 'school_admins')) WITH CHECK (permission_check(resource_path, 'school_admins'));
-- ALTER TABLE public.school_admins ENABLE ROW LEVEL security;
-- ALTER TABLE public.school_admins FORCE ROW LEVEL security;

-- Clone granted_role table from usermgmt
CREATE TABLE IF NOT EXISTS public.granted_role (
    granted_role_id TEXT NOT NULL UNIQUE,
    user_group_id TEXT NOT NULL,
    role_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id)
    -- CONSTRAINT fk__granted_role__user_group_id FOREIGN KEY (user_group_id) REFERENCES public.user_group(user_group_id),
    -- CONSTRAINT fk__granted_role__role_id FOREIGN KEY (role_id) REFERENCES public.role(role_id)
);
CREATE POLICY rls_granted_role ON public.granted_role USING (permission_check(resource_path, 'granted_role')) WITH CHECK (permission_check(resource_path, 'granted_role'));
ALTER TABLE IF EXISTS public.granted_role ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.granted_role FORCE ROW LEVEL security;

-- Clone user_group_member table from usermgmt
CREATE TABLE IF NOT EXISTS public.user_group_member (
    user_id TEXT NOT NULL,
    user_group_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id)
    -- CONSTRAINT fk__user_group_member__user_id FOREIGN KEY (user_id) REFERENCES public.users(user_id),
    -- CONSTRAINT fk__user_group_member__user_group_id FOREIGN KEY (user_group_id) REFERENCES public.user_group(user_group_id)
);
CREATE POLICY rls_user_group_member ON public.user_group_member USING (permission_check(resource_path, 'user_group_member')) with check (permission_check(resource_path, 'user_group_member'));
ALTER TABLE IF EXISTS public.user_group_member ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.user_group_member FORCE ROW LEVEL security;

-- Enable autofillresourcepath
ALTER TABLE IF EXISTS public.role ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE IF EXISTS public.user_group ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE IF EXISTS public.users_groups ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE IF EXISTS public.groups ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
-- ALTER TABLE public.school_admins ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE IF EXISTS public.granted_role ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE IF EXISTS public.user_group_member ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();

-- Add missing users columns
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS country text NOT NULL;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS avatar text;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS phone_number text;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS email text;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS device_token text;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS allow_notification boolean;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone NOT NULL;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS created_at timestamp with time zone NOT NULL;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS is_tester boolean;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS facebook_id text;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS platform text;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS phone_verified boolean;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS email_verified boolean;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS given_name TEXT;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS last_login_date timestamp with time zone;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS birthday DATE NULL;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS gender TEXT;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '';
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '';
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS first_name_phonetic TEXT;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS last_name_phonetic TEXT;
ALTER TABLE IF EXISTS ONLY public.users ADD COLUMN IF NOT EXISTS full_name_phonetic TEXT;

-- Update not-null constraints for cloned tables
ALTER TABLE IF EXISTS public.user_group_member ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE IF EXISTS public.granted_role ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE IF EXISTS public.role ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE IF EXISTS public.user_group ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE IF EXISTS public.users_groups ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE IF EXISTS public.groups ALTER COLUMN resource_path SET NOT NULL;
-- ALTER TABLE public.school_admins ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE IF EXISTS public.locations ALTER COLUMN name SET NOT NULL;

ALTER TABLE IF EXISTS public.role ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;
ALTER TABLE IF EXISTS public.user_group ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;