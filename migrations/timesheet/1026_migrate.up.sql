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
ALTER TABLE public.role ENABLE ROW LEVEL security;
ALTER TABLE public.role FORCE ROW LEVEL security;

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
ALTER TABLE public.user_group ENABLE ROW LEVEL security;
ALTER TABLE public.user_group FORCE ROW LEVEL security;

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
);
CREATE POLICY rls_granted_role ON public.granted_role USING (permission_check(resource_path, 'granted_role')) WITH CHECK (permission_check(resource_path, 'granted_role'));
ALTER TABLE public.granted_role ENABLE ROW LEVEL security;
ALTER TABLE public.granted_role FORCE ROW LEVEL security;

-- Clone user_group_member table from usermgmt
CREATE TABLE IF NOT EXISTS public.user_group_member (
    user_id TEXT NOT NULL,
    user_group_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id)
);
CREATE POLICY rls_user_group_member ON public.user_group_member USING (permission_check(resource_path, 'user_group_member')) with check (permission_check(resource_path, 'user_group_member'));
ALTER TABLE public.user_group_member ENABLE ROW LEVEL security;
ALTER TABLE public.user_group_member FORCE ROW LEVEL security;

-- Enable autofillresourcepath
ALTER TABLE public.role ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE public.user_group ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE public.granted_role ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();
ALTER TABLE public.user_group_member ALTER COLUMN resource_path SET DEFAULT autofillresourcepath();

-- Add missing users columns
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS country text NOT NULL;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS avatar text;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS phone_number text;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS email text;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS device_token text;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS allow_notification boolean;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS updated_at timestamp with time zone NOT NULL;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS created_at timestamp with time zone NOT NULL;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS is_tester boolean;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS facebook_id text;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS platform text;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS phone_verified boolean;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS email_verified boolean;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS given_name TEXT;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS last_login_date timestamp with time zone;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS birthday DATE NULL;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS gender TEXT;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '';
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '';
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS first_name_phonetic TEXT;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS last_name_phonetic TEXT;
ALTER TABLE ONLY public.users ADD COLUMN IF NOT EXISTS full_name_phonetic TEXT;

-- Update not-null constraints for cloned tables
ALTER TABLE public.user_group_member ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE public.granted_role ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE public.role ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE public.user_group ALTER COLUMN resource_path SET NOT NULL;
ALTER TABLE public.locations ALTER COLUMN name SET NOT NULL;

-- Addd missing columns
ALTER TABLE public.role ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;
ALTER TABLE public.user_group ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;
ALTER TABLE public.user_group ADD COLUMN IF NOT EXISTS org_location_id TEXT;
