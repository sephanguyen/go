CREATE TABLE public.users (
    user_id text NOT NULL,
    name text NOT NULL,
    user_group text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT users_pk PRIMARY KEY (user_id)
);

CREATE POLICY rls_users ON "users" USING (permission_check(resource_path, 'users')) WITH CHECK (permission_check(resource_path, 'users'));

ALTER TABLE "users" ENABLE ROW LEVEL security;
ALTER TABLE "users" FORCE ROW LEVEL security;
