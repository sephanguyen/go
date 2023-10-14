CREATE TABLE IF NOT EXISTS public.file(
    file_id text NOT NULL,
    file_name text NOT NULL,
    file_type text NOT NULL,
    download_link text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath()
);

CREATE POLICY rls_file ON public.file
    USING (permission_check(resource_path, 'file'))
    WITH CHECK (permission_check(resource_path, 'file'));

CREATE POLICY rls_file_restrictive ON "file"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'file'))
    WITH CHECK (permission_check(resource_path, 'file'));
ALTER TABLE public.file ENABLE ROW LEVEL security;
ALTER TABLE public.file FORCE ROW LEVEL security;