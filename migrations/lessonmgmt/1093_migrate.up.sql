CREATE TABLE public.classes (
	class_id serial4 NOT NULL,
	school_id serial4 NOT NULL,
	avatar text NOT NULL,
	"name" text NOT NULL,
	subjects _text NULL,
	grades _int4 NULL,
	status text NOT NULL DEFAULT 'CLASS_STATUS_NONE'::text,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	plan_id text NULL,
	country text NULL,
	plan_expired_at timestamptz NULL,
	plan_duration int2 NULL,
	class_code text NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT classes_pk PRIMARY KEY (class_id),
	CONSTRAINT classes_un UNIQUE (class_code)
);

CREATE POLICY rls_classes ON "classes" USING (permission_check(resource_path, 'classes'::text)) WITH CHECK (permission_check(resource_path, 'classes'::text));
CREATE POLICY rls_classes_restrictive ON "classes" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'classes'::text)) WITH CHECK (permission_check(resource_path, 'classes'::text));

ALTER TABLE "classes" ENABLE ROW LEVEL security;
ALTER TABLE "classes" FORCE ROW LEVEL security;