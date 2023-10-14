CREATE TABLE IF NOT EXISTS public.user_group_member (
    user_id TEXT NOT NULL,
    user_group_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id)
);
CREATE POLICY rls_user_group_member ON "user_group_member" USING (permission_check(resource_path, 'user_group_member')) WITH CHECK (permission_check(resource_path, 'user_group_member'));
CREATE POLICY rls_user_group_member_restrictive ON "user_group_member" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'user_group_member')) with check (permission_check(resource_path, 'user_group_member'));

ALTER TABLE "user_group_member" ENABLE ROW LEVEL security;
ALTER TABLE "user_group_member" FORCE ROW LEVEL security;
