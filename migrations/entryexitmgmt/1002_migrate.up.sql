CREATE TABLE IF NOT EXISTS public.users (
    user_id text UNIQUE NOT NULL,
    country text NOT NULL,
    name text NOT NULL,
    given_name text,
    device_token text,
    allow_notification boolean,
    user_group text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT users_pk PRIMARY KEY (user_id),
    CONSTRAINT users_user_group_check CHECK ((user_group = ANY (ARRAY['USER_GROUP_STUDENT'::text, 'USER_GROUP_COACH'::text, 'USER_GROUP_TUTOR'::text, 'USER_GROUP_STAFF'::text, 'USER_GROUP_ADMIN'::text, 'USER_GROUP_TEACHER'::text, 'USER_GROUP_PARENT'::text, 'USER_GROUP_CONTENT_ADMIN'::text, 'USER_GROUP_CONTENT_STAFF'::text, 'USER_GROUP_SALES_ADMIN'::text, 'USER_GROUP_SALES_STAFF'::text, 'USER_GROUP_CS_ADMIN'::text, 'USER_GROUP_CS_STAFF'::text, 'USER_GROUP_SCHOOL_ADMIN'::text, 'USER_GROUP_SCHOOL_STAFF'::text])))
);

CREATE POLICY rls_users ON "users" USING (permission_check(resource_path, 'users')) WITH CHECK (permission_check(resource_path, 'users'));

ALTER TABLE "users" ENABLE ROW LEVEL security;
ALTER TABLE "users" FORCE ROW LEVEL security;
