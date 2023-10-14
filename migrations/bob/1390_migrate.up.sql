CREATE TABLE IF NOT EXISTS public.package (
    package_id TEXT NOT NULL,
    package_type TEXT NOT NULL,
    max_slot INTEGER NOT NULL,
    package_start_date TIMESTAMP WITH TIME ZONE,
    package_end_date   TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT package_pk PRIMARY KEY (package_id)
 );

CREATE POLICY rls_package ON "package"
USING (permission_check(resource_path, 'package')) WITH CHECK (permission_check(resource_path, 'package'));
CREATE POLICY rls_package_restrictive ON "package" AS RESTRICTIVE
USING (permission_check(resource_path, 'package')) WITH CHECK (permission_check(resource_path, 'package'));

ALTER TABLE "package" ENABLE ROW LEVEL security;
ALTER TABLE "package" FORCE ROW LEVEL security;
