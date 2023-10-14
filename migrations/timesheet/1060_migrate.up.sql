CREATE TABLE IF NOT EXISTS timesheet_confirmation_cut_off_date
(
    id                    TEXT NOT NULL,
    cut_off_date          INTEGER NOT NULL,
    start_date            TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date              TIMESTAMP WITH TIME ZONE,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at            TIMESTAMP WITH TIME ZONE,
    resource_path         TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT timesheet_confirmation_cut_off_date__id__pk
        PRIMARY KEY (id)
);

CREATE POLICY rls_timesheet_confirmation_cut_off_date ON "timesheet_confirmation_cut_off_date"
    USING (permission_check (resource_path, 'timesheet_confirmation_cut_off_date'))
    WITH CHECK (permission_check (resource_path, 'timesheet_confirmation_cut_off_date'));

CREATE POLICY rls_timesheet_confirmation_cut_off_date_restrictive ON "timesheet_confirmation_cut_off_date" AS RESTRICTIVE FOR ALL TO PUBLIC 
USING (permission_check(resource_path, 'timesheet_confirmation_cut_off_date'))
WITH CHECK (permission_check(resource_path, 'timesheet_confirmation_cut_off_date'));

ALTER TABLE "timesheet_confirmation_cut_off_date" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "timesheet_confirmation_cut_off_date" FORCE ROW LEVEL SECURITY;


CREATE TABLE IF NOT EXISTS timesheet_confirmation_period
(
    id                    TEXT NOT NULL,
    start_date            TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date              TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at            TIMESTAMP WITH TIME ZONE,
    resource_path         TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT timesheet_confirmation_period__id__pk
        PRIMARY KEY (id)
);

CREATE POLICY rls_timesheet_confirmation_period ON "timesheet_confirmation_period"
    USING (permission_check (resource_path, 'timesheet_confirmation_period'))
    WITH CHECK (permission_check (resource_path, 'timesheet_confirmation_period'));

CREATE POLICY rls_timesheet_confirmation_period_restrictive ON "timesheet_confirmation_period" AS RESTRICTIVE FOR ALL TO PUBLIC 
USING (permission_check(resource_path, 'timesheet_confirmation_period'))
WITH CHECK (permission_check(resource_path, 'timesheet_confirmation_period'));

ALTER TABLE "timesheet_confirmation_period" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "timesheet_confirmation_period" FORCE ROW LEVEL SECURITY;


CREATE TABLE IF NOT EXISTS timesheet_confirmation_info
(
    id                    TEXT NOT NULL,
    period_id             TEXT NOT NULL,
    location_id           TEXT NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at            TIMESTAMP WITH TIME ZONE,
    resource_path         TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT timesheet_confirmation_info__id__pk
        PRIMARY KEY (id),
    CONSTRAINT timesheet_confirmation_info__period_id__fk
        FOREIGN KEY (period_id) REFERENCES public.timesheet_confirmation_period(id),
    CONSTRAINT timesheet_confirmation_info__location_id__fk
        FOREIGN KEY (location_id) REFERENCES public.locations(location_id)
);

CREATE POLICY rls_timesheet_confirmation_info ON "timesheet_confirmation_info"
    USING (permission_check (resource_path, 'timesheet_confirmation_info'))
    WITH CHECK (permission_check (resource_path, 'timesheet_confirmation_info'));

CREATE POLICY rls_timesheet_confirmation_info_restrictive ON "timesheet_confirmation_info" AS RESTRICTIVE FOR ALL TO PUBLIC 
USING (permission_check(resource_path, 'timesheet_confirmation_info'))
WITH CHECK (permission_check(resource_path, 'timesheet_confirmation_info'));

ALTER TABLE "timesheet_confirmation_info" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "timesheet_confirmation_info" FORCE ROW LEVEL SECURITY;