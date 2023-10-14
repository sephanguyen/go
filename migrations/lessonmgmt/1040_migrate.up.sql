CREATE TABLE staff (
	staff_id text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	auto_create_timesheet bool NULL DEFAULT false,
	working_status text NOT NULL DEFAULT 'AVAILABLE'::text,
	start_date date NULL,
	end_date date NULL,
	CONSTRAINT pk__staff PRIMARY KEY (staff_id)
);
CREATE INDEX staff__created_at__idx_desc ON public.staff USING btree (created_at DESC);
CREATE INDEX staff__staff_id__idx ON public.staff USING btree (resource_path);

CREATE POLICY rls_staff ON "staff"
USING (permission_check(resource_path, 'staff'))
WITH CHECK (permission_check(resource_path, 'staff'));

CREATE POLICY rls_staff_restrictive ON "staff" AS RESTRICTIVE FOR ALL TO PUBLIC 
USING (permission_check(resource_path, 'staff'::text))
WITH CHECK (permission_check(resource_path, 'staff'::text));

ALTER TABLE "staff" ENABLE ROW LEVEL security;
ALTER TABLE "staff" FORCE ROW LEVEL security;