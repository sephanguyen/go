CREATE TABLE public.class_members (
	class_member_id text NOT NULL,
	class_id int4 NOT NULL,
	user_id text NOT NULL,
	status text NOT NULL DEFAULT 'CLASS_MEMBER_STATUS_NONE'::text,
	user_group text NOT NULL,
	is_owner bool NOT NULL DEFAULT false,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	student_subscription_id text NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT class_members_pk PRIMARY KEY (class_member_id)
);

CREATE INDEX class_members_user_id_idx ON public.class_members USING btree (user_id);

CREATE POLICY rls_class_members ON "class_members" USING (permission_check(resource_path, 'class_members'::text)) WITH CHECK (permission_check(resource_path, 'class_members'::text));
CREATE POLICY rls_class_members_restrictive ON "class_members" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'class_members'::text)) WITH CHECK (permission_check(resource_path, 'class_members'::text));

ALTER TABLE "class_members" ENABLE ROW LEVEL security;
ALTER TABLE "class_members" FORCE ROW LEVEL security;