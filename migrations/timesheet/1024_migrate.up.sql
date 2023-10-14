CREATE TABLE IF NOT EXISTS public.transportation_expense(
    transportation_expense_id TEXT NOT NULL,
    timesheet_id TEXT NOT NULL,
    transportation_type TEXT NOT NULL,
    transportation_from TEXT NOT NULL,
    transportation_to TEXT NOT NULL,
    cost_amount NUMERIC(12,2),
    round_trip BOOLEAN NOT NULL,
    remarks TEXT NOT NULL,
    created_at              TIMESTAMP with time zone NOT NULL,
    updated_at              TIMESTAMP with time zone NOT NULL,
    deleted_at              TIMESTAMP with time zone,
    resource_path           TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT transportation_expense__pk
            PRIMARY KEY (transportation_expense_id),
    CONSTRAINT transportation_expense_timesheet_id__fk
        FOREIGN KEY (timesheet_id) REFERENCES public.timesheet(timesheet_id)
);

CREATE POLICY rls_transportation_expense ON transportation_expense USING (permission_check(resource_path, 'transportation_expense'))
WITH CHECK (permission_check(resource_path, 'transportation_expense'));

ALTER TABLE "transportation_expense"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "transportation_expense"
    FORCE ROW LEVEL SECURITY;