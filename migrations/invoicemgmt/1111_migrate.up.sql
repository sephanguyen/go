CREATE TABLE IF NOT EXISTS public.company_detail (
    company_detail_id TEXT NOT NULL,
    company_name TEXT NOT NULL,
    company_address TEXT,
    company_phone_number TEXT,
    company_logo_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__company_detail PRIMARY KEY (company_detail_id)
);

CREATE POLICY rls_company_detail ON "company_detail"
USING (permission_check(resource_path, 'company_detail')) WITH CHECK (permission_check(resource_path, 'company_detail'));

CREATE POLICY rls_company_detail_restrictive ON "company_detail" AS RESTRICTIVE
USING (permission_check(resource_path, 'company_detail'))WITH CHECK (permission_check(resource_path, 'company_detail'));

ALTER TABLE "company_detail" ENABLE ROW LEVEL security;
ALTER TABLE "company_detail" FORCE ROW LEVEL security;