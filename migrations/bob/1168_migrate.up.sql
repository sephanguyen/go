CREATE TABLE IF NOT EXISTS public.mastermgmt_import_log (
    mastermgmt_import_log_id TEXT NOT NULL PRIMARY KEY,
    user_id text NOT NULL,
    import_type text NOT NULL,
    payload jsonb,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath()
);

CREATE POLICY rls_mastermgmt_import_log ON "mastermgmt_import_log" using (permission_check(resource_path, 'mastermgmt_import_log')) with check (permission_check(resource_path, 'mastermgmt_import_log'));

ALTER TABLE "mastermgmt_import_log" ENABLE ROW LEVEL security;
ALTER TABLE "mastermgmt_import_log" FORCE ROW LEVEL security; 
