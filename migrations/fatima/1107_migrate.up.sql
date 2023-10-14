CREATE TABLE IF NOT EXISTS public.user_tag(
    user_tag_id text NOT NULL,
    user_tag_name text NOT NULL,
    user_tag_type text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL,
    CONSTRAINT pk__user_tag PRIMARY KEY (user_tag_id)
);

CREATE POLICY rls_user_tag ON public.user_tag
USING (permission_check(resource_path, 'user_tag'))
WITH CHECK (permission_check(resource_path, 'user_tag'));

CREATE TABLE IF NOT EXISTS public.tagged_user (
    user_id text NOT NULL,
    tag_id text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text,
    CONSTRAINT pk__tagged_user PRIMARY KEY (user_id, tag_id)
);
CREATE POLICY rls_tagged_user ON public.tagged_user
USING (permission_check(resource_path, 'tagged_user'))
WITH CHECK (permission_check(resource_path, 'tagged_user'));
