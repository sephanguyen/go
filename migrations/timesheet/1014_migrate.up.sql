CREATE TABLE IF NOT EXISTS public.other_working_hours (
    other_working_hours_id  TEXT NOT NULL CONSTRAINT other_working_hours__pk PRIMARY KEY,
    timesheet_id            TEXT NOT NULL,
    timesheet_config_id     TEXT NOT NULL,
    start_time              TIMESTAMP with time zone NOT NULL,
    end_time                TIMESTAMP with time zone NOT NULL,
    total_hour              SMALLINT NOT NULL,
    content                 TEXT,
    created_at              TIMESTAMP with time zone NOT NULL,
    updated_at              TIMESTAMP with time zone NOT NULL,
    deleted_at              TIMESTAMP with time zone,
    resource_path           TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT fk__other_working_hours__timesheet__timesheet_id 
        FOREIGN KEY (timesheet_id) REFERENCES public.timesheet(timesheet_id),
    CONSTRAINT fk__other_working_hours__timesheet_config__timesheet_config_id 
        FOREIGN KEY (timesheet_config_id) REFERENCES public.timesheet_config(timesheet_config_id)
);

CREATE POLICY rls_other_working_hours ON "other_working_hours"
    USING (permission_check (resource_path, 'other_working_hours'))
    WITH CHECK (permission_check (resource_path, 'other_working_hours'));

ALTER TABLE "other_working_hours" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "other_working_hours" FORCE ROW LEVEL SECURITY;