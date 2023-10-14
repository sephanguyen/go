CREATE TABLE IF NOT EXISTS auto_create_timesheet_flag
(
    staff_id              TEXT NOT NULL,
    flag_on               boolean NOT NULL DEFAULT false,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at            TIMESTAMP WITH TIME ZONE,
    resource_path         TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT auto_create_timesheet_flag__pk
        PRIMARY KEY (staff_id),
    CONSTRAINT auto_create_timesheet_flag_staff_id__fk
        FOREIGN KEY (staff_id) REFERENCES public.staff(staff_id)
);

CREATE POLICY rls_auto_create_timesheet_flag ON "auto_create_timesheet_flag"
    USING (permission_check (resource_path, 'auto_create_timesheet_flag'))
    WITH CHECK (permission_check (resource_path, 'auto_create_timesheet_flag'));

ALTER TABLE "auto_create_timesheet_flag" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "auto_create_timesheet_flag" FORCE ROW LEVEL SECURITY;