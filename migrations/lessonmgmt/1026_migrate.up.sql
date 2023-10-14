CREATE TABLE IF NOT EXISTS public.classroom (
    classroom_id TEXT NOT NULL,
    name TEXT NOT NULL,
    location_id TEXT NOT NULL,
    remarks TEXT NULL,
    is_archived boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT classroom_pk PRIMARY KEY (classroom_id)
);

CREATE POLICY rls_classroom ON "classroom" using (permission_check(resource_path, 'classroom')) with check (permission_check(resource_path, 'classroom'));
CREATE POLICY rls_classroom_restrictive ON "classroom" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'classroom')) with check (permission_check(resource_path, 'classroom'));

ALTER TABLE "classroom" ENABLE ROW LEVEL security;
ALTER TABLE "classroom" FORCE ROW LEVEL security;
