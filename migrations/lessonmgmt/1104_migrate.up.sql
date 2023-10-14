CREATE TABLE IF NOT EXISTS public.classdo_account (
    classdo_id TEXT NOT NULL,
    classdo_email TEXT NOT NULL,
    classdo_api_key TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT pk__classdo_account PRIMARY KEY (classdo_id)
);

CREATE INDEX classdo_account_classdo_email_idx on classdo_account using HASH("classdo_email");

CREATE POLICY rls_classdo_account ON "classdo_account" USING (permission_check(resource_path, 'classdo_account')) WITH CHECK (permission_check(resource_path, 'classdo_account'));
CREATE POLICY rls_classdo_account_restrictive ON "classdo_account" AS RESTRICTIVE USING (permission_check(resource_path, 'classdo_account'))WITH CHECK (permission_check(resource_path, 'classdo_account'));

ALTER TABLE "classdo_account" ENABLE ROW LEVEL security;
ALTER TABLE "classdo_account" FORCE ROW LEVEL security;