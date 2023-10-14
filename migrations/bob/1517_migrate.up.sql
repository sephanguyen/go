CREATE TABLE IF NOT EXISTS tmp_notification_assignment_data (
    notification_id TEXT PRIMARY KEY, 
    data JSONB,
    resource_path text DEFAULT autofillresourcepath() NOT NULL
);

DROP POLICY IF EXISTS rls_tmp_notification_assignment_data ON "tmp_notification_assignment_data";
CREATE POLICY rls_tmp_notification_assignment_data ON "tmp_notification_assignment_data"
    USING (permission_check(resource_path, 'tmp_notification_assignment_data'))
    WITH CHECK (permission_check(resource_path, 'tmp_notification_assignment_data'));

DROP POLICY IF EXISTS rls_tmp_notification_assignment_data_restrictive ON "tmp_notification_assignment_data";
CREATE POLICY rls_tmp_notification_assignment_data_restrictive ON "tmp_notification_assignment_data"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'tmp_notification_assignment_data'))
    WITH CHECK (permission_check(resource_path, 'tmp_notification_assignment_data'));

ALTER TABLE "tmp_notification_assignment_data" ENABLE ROW LEVEL security;
ALTER TABLE "tmp_notification_assignment_data" FORCE ROW LEVEL security;
