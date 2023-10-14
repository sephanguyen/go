create table user_tag
(
    user_tag_id text not null,
    user_tag_name text not null,
    user_tag_type text not null,
    is_archived boolean not null,
    created_at timestamp with time zone default timezone('utc'::text, now()) not null,
    updated_at timestamp with time zone default timezone('utc'::text, now()) not null,
    deleted_at timestamp with time zone,
    resource_path text default autofillresourcepath() not null,
    user_tag_partner_id text not null,

    constraint user_tag_pk primary key (user_tag_id)
);

CREATE POLICY rls_user_tag ON "user_tag" USING (
    permission_check(resource_path, 'user_tag'))
WITH CHECK (
    permission_check(resource_path, 'user_tag'));

CREATE POLICY rls_user_tag_restrictive ON "user_tag" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'user_tag'))
WITH CHECK (
    permission_check(resource_path, 'user_tag'));

ALTER TABLE "user_tag" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "user_tag" FORCE ROW LEVEL SECURITY;

create table tagged_user
(
    user_id text not null,
    tag_id text not null,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null,
    deleted_at timestamp with time zone,
    resource_path text,

    constraint tagged_user_pk primary key (user_id, tag_id)
);

CREATE POLICY rls_tagged_user ON "tagged_user" USING (
    permission_check(resource_path, 'tagged_user'))
WITH CHECK (
    permission_check(resource_path, 'tagged_user'));

CREATE POLICY rls_tagged_user_restrictive ON "tagged_user" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'tagged_user'))
WITH CHECK (
    permission_check(resource_path, 'tagged_user'));

ALTER TABLE "tagged_user" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "tagged_user" FORCE ROW LEVEL SECURITY;
