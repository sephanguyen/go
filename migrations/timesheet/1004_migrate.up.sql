CREATE TABLE IF NOT EXISTS public.staff
(
    staff_id      TEXT                     NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    UPDATED_AT    TIMESTAMP WITH TIME ZONE NOT NULL,
    DELETED_AT    TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT pk__staff PRIMARY KEY (staff_id)
);
CREATE POLICY rls_staff ON "staff" USING (permission_check(resource_path, 'staff')) WITH CHECK (permission_check(resource_path, 'staff'));
ALTER TABLE "staff"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "staff"
    FORCE ROW LEVEL SECURITY;