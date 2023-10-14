CREATE TABLE IF NOT EXISTS public.user_device_tokens (
	user_device_token_id serial4 NOT NULL,
	user_id text NOT NULL,
	device_token text NULL,
	allow_notification bool NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
	CONSTRAINT user_device_tokens_pk PRIMARY KEY (user_device_token_id),
	CONSTRAINT user_id_un UNIQUE (user_id)
);

CREATE POLICY rls_user_device_tokens ON "user_device_tokens"
USING (permission_check(resource_path, 'user_device_tokens'))
WITH CHECK (permission_check(resource_path, 'user_device_tokens'));

ALTER TABLE "user_device_tokens" ENABLE ROW LEVEL security;
ALTER TABLE "user_device_tokens" FORCE ROW LEVEL security;