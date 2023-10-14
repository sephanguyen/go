CREATE TABLE IF NOT EXISTS public.timesheet_config (
    timesheet_config_id TEXT NOT NULL CONSTRAINT timesheet_config__pk PRIMARY KEY,
    config_key          TEXT NOT NULL,
    config_value        TEXT NOT NULL,
    created_at          TIMESTAMP with time zone NOT NULL,
    updated_at          TIMESTAMP with time zone NOT NULL,
    deleted_at          TIMESTAMP with time zone,
    resource_path       TEXT DEFAULT autofillresourcepath()
);

CREATE POLICY rls_timesheet_config ON "timesheet_config"
    USING (permission_check (resource_path, 'timesheet_config'))
    WITH CHECK (permission_check (resource_path, 'timesheet_config'));

ALTER TABLE "timesheet_config" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "timesheet_config" FORCE ROW LEVEL SECURITY;