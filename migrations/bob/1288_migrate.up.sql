-- add constrains for existed table user_tag
DELETE FROM user_tag WHERE user_tag_type = 'USER_TAG_TYPE_NONE';
ALTER TABLE ONLY public.user_tag DROP CONSTRAINT IF EXISTS user_tag_user_tag_type_check;
ALTER TABLE public.user_tag ADD CONSTRAINT user_tag__user_tag_type__check CHECK (
  user_tag_type = ANY (ARRAY[
		'USER_TAG_TYPE_STUDENT',
		'USER_TAG_TYPE_STUDENT_DISCOUNT',
		'USER_TAG_TYPE_PARENT',
		'USER_TAG_TYPE_PARENT_DISCOUNT'
  ]::text[])
);

-- create table tagged_user
CREATE TABLE IF NOT EXISTS public.tagged_user (
    user_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT,

    CONSTRAINT pk__tagged_user PRIMARY KEY (user_id, tag_id),
    CONSTRAINT fk__tagged_user__user_id FOREIGN KEY (user_id) REFERENCES public.users(user_id),
    CONSTRAINT fk__tagged_user__tag_id FOREIGN KEY (tag_id) REFERENCES public.user_tag(user_tag_id)
);

CREATE POLICY rls_tagged_user ON public.tagged_user
USING (permission_check(resource_path, 'tagged_user'))
WITH CHECK (permission_check(resource_path, 'tagged_user'));

CREATE POLICY rls_tagged_user_restrictive ON "tagged_user" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'tagged_user'))
with check (permission_check(resource_path, 'tagged_user'));

ALTER TABLE public.tagged_user ENABLE ROW LEVEL security;
ALTER TABLE public.tagged_user FORCE ROW LEVEL security;
