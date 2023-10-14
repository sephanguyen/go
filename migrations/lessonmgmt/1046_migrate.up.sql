CREATE TABLE IF NOT EXISTS public.zoom_account (
    zoom_id TEXT NOT NULL,
    email TEXT NOT NULL,
    user_name TEXT,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT pk__zoom_account PRIMARY KEY (zoom_id)
);

CREATE INDEX zoom_account_email_idx on zoom_account using HASH("email");

CREATE POLICY rls_zoom_account ON "zoom_account"
USING (permission_check(resource_path, 'zoom_account')) WITH CHECK (permission_check(resource_path, 'zoom_account'));
CREATE POLICY rls_zoom_account_restrictive ON "zoom_account" AS RESTRICTIVE
USING (permission_check(resource_path, 'zoom_account'))WITH CHECK (permission_check(resource_path, 'zoom_account'));

ALTER TABLE "zoom_account" ENABLE ROW LEVEL security;
ALTER TABLE "zoom_account" FORCE ROW LEVEL security;

ALTER TABLE public.lessons
    ADD COLUMN IF NOT EXISTS "zoom_link" TEXT,
    ADD COLUMN IF NOT EXISTS "zoom_owner_id" TEXT,
    ADD COLUMN IF NOT EXISTS "zoom_id" TEXT,
    ADD COLUMN IF NOT EXISTS "zoom_occurrence_id" TEXT;

