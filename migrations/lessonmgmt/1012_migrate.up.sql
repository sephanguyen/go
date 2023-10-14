-- public.media definition

CREATE TABLE public.media (
	media_id text NOT NULL,
	name text NULL,
	resource text NULL,
	"comments" jsonb NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	"type" text NULL,
	converted_images jsonb NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	file_size_bytes int8 NULL DEFAULT 0,
	duration_seconds int4 NULL DEFAULT 0,
	CONSTRAINT media_pk PRIMARY KEY (media_id)
);

CREATE POLICY rls_media ON "media" USING (permission_check(resource_path, 'media')) WITH CHECK (permission_check(resource_path, 'media'));
CREATE POLICY rls_media_restrictive ON "media" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'media')) with check (permission_check(resource_path, 'media'));

ALTER TABLE "media" ENABLE ROW LEVEL security;
ALTER TABLE "media" FORCE ROW LEVEL security;