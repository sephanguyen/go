CREATE TABLE IF NOT EXISTS partner_auto_create_timesheet_flag
(
    id                    TEXT NOT NULL,
    flag_on               BOOLEAN NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at            TIMESTAMP WITH TIME ZONE,
    resource_path         TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT partner_auto_create_timesheet_flag__id__pk
        PRIMARY KEY (id),
    CONSTRAINT partner_auto_create_timesheet_flag_resource_unique
        UNIQUE (resource_path)
);

CREATE POLICY rls_partner_auto_create_timesheet_flag ON "partner_auto_create_timesheet_flag"
    USING (permission_check (resource_path, 'partner_auto_create_timesheet_flag'))
    WITH CHECK (permission_check (resource_path, 'partner_auto_create_timesheet_flag'));

CREATE POLICY rls_partner_auto_create_timesheet_flag_restrictive ON "partner_auto_create_timesheet_flag" AS RESTRICTIVE FOR ALL TO PUBLIC 
USING (permission_check(resource_path, 'partner_auto_create_timesheet_flag'))
WITH CHECK (permission_check(resource_path, 'partner_auto_create_timesheet_flag'));

ALTER TABLE "partner_auto_create_timesheet_flag" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "partner_auto_create_timesheet_flag" FORCE ROW LEVEL SECURITY;