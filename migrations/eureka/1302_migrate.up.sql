CREATE TABLE IF NOT EXISTS public.content_bank_medias (
    id text NOT NULL,
    name text NOT NULL,
    resource text,
    type text,
    file_size_bytes int8 NULL DEFAULT 0,
    created_by text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath() NOT NULL
);

CREATE POLICY rls_content_bank_medias ON "content_bank_medias" using (
  permission_check(resource_path, 'content_bank_medias')
) with check (
  permission_check(resource_path, 'content_bank_medias')
);

CREATE POLICY rls_content_bank_medias_restrictive ON "content_bank_medias" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'content_bank_medias')
) with check (
    permission_check(resource_path, 'content_bank_medias')
);

ALTER TABLE "content_bank_medias" ENABLE ROW LEVEL security;
ALTER TABLE "content_bank_medias" FORCE ROW LEVEL security;

CREATE UNIQUE INDEX IF NOT EXISTS content_bank_medias_name_resource_path_unique_idx ON public.content_bank_medias (name,resource_path) WHERE (deleted_at IS NULL);

ALTER TABLE ONLY public.content_bank_medias DROP CONSTRAINT IF EXISTS media_type_check;
ALTER TABLE public.content_bank_medias
    ADD CONSTRAINT media_type_check CHECK (type = ANY (ARRAY[
                'MEDIA_TYPE_IMAGE'::text
            ]));
