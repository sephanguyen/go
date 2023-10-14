CREATE TABLE IF NOT EXISTS public.staff (
    staff_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT pk__staff PRIMARY KEY (staff_id),
    CONSTRAINT fk__staff__staff_id FOREIGN KEY (staff_id)
                REFERENCES public.users(user_id)
);

CREATE POLICY rls_staff ON "staff"
USING (permission_check(resource_path, 'staff'))
WITH CHECK (permission_check(resource_path, 'staff'));

ALTER TABLE "staff" ENABLE ROW LEVEL security;
ALTER TABLE "staff" FORCE ROW LEVEL security;
