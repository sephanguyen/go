CREATE TABLE IF NOT EXISTS public.timesheet
(
    timesheet_id      TEXT                     NOT NULL
        CONSTRAINT timesheet__pk PRIMARY KEY,
    staff_id          TEXT NOT NULL,
    location_id       TEXT NOT NULL,
    time_sheet_status TEXT NOT NULL,
    timesheet_date    TIMESTAMP WITH TIME ZONE NOT NULL,
    remark            TEXT,
    resource_path     TEXT DEFAULT autofillresourcepath(),
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at        TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk__staff__staff_id FOREIGN KEY (staff_id) REFERENCES public.staff(staff_id),
    CONSTRAINT fk__location__location_id FOREIGN KEY (location_id) REFERENCES public.locations(location_id)
);
CREATE POLICY rls_timesheet ON "timesheet" USING (permission_check(resource_path, 'timesheet')) WITH CHECK (permission_check(resource_path, 'timesheet'));
ALTER TABLE "timesheet"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "timesheet"
    FORCE ROW LEVEL SECURITY;