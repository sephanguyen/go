CREATE TABLE IF NOT EXISTS public.staff_transportation_expense (
    id          TEXT NOT NULL,
    staff_id    TEXT NOT NULL,
    location_id TEXT NOT NULL,
    transportation_type TEXT NOT NULL,
    transportation_from TEXT NOT NULL,
    transportation_to   TEXT NOT NULL,
    cost_amount     integer DEFAULT 0,
    round_trip      BOOLEAN NOT NULL,
    remarks         TEXT NOT NULL,

    created_at      TIMESTAMP with time zone NOT NULL,
    updated_at      TIMESTAMP with time zone NOT NULL,
    deleted_at      TIMESTAMP with time zone,
    resource_path   TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT staff_transportation_expense__id__pk
        PRIMARY KEY (id),
    CONSTRAINT staff_transportation_expense__staff_id__fk
        FOREIGN KEY (staff_id) REFERENCES public.staff(staff_id),
    CONSTRAINT staff_transportation_expense__location_id__fk
        FOREIGN KEY (location_id) REFERENCES public.locations(location_id)
);

CREATE POLICY rls_staff_transportation_expense ON staff_transportation_expense 
USING (permission_check(resource_path, 'staff_transportation_expense'))
WITH CHECK (permission_check(resource_path, 'staff_transportation_expense'));

CREATE POLICY rls_staff_transportation_expense_restrictive ON "staff_transportation_expense" AS RESTRICTIVE FOR ALL TO PUBLIC 
USING (permission_check(resource_path, 'staff_transportation_expense'))
WITH CHECK (permission_check(resource_path, 'staff_transportation_expense'));

ALTER TABLE "staff_transportation_expense"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "staff_transportation_expense"
    FORCE ROW LEVEL SECURITY;