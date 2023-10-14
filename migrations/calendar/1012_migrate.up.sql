CREATE TABLE IF NOT EXISTS public.timetable (
	timetable_id text,
	parent_id text,
	location_id text,
	scheduling_type text,

	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,

	CONSTRAINT timetable_pk PRIMARY KEY (timetable_id)
);

ALTER TABLE public.timetable ADD CONSTRAINT timetable_fk FOREIGN KEY (parent_id) REFERENCES public.timetable(timetable_id);

CREATE POLICY rls_timetable ON "timetable" using (permission_check(resource_path, 'timetable')) with check (permission_check(resource_path, 'timetable'));
CREATE POLICY rls_timetable_restrictive ON "timetable" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'timetable')) with check (permission_check(resource_path, 'timetable'));

ALTER TABLE "timetable" ENABLE ROW LEVEL security;
ALTER TABLE "timetable" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.scheduling_slot (
	scheduling_slot_id text NOT NULL,
	timetable_id text NOT NULL,
	student_id text NOT NULL,
	teacher_id text NOT NULL,
	course_id text NOT NULL,
	classroom_id text NOT NULL,
	is_confirmed bool DEFAULT FALSE,
   
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,

	CONSTRAINT scheduling_slot_pk PRIMARY KEY (scheduling_slot_id)
);

ALTER TABLE public.scheduling_slot ADD CONSTRAINT scheduling_slot_fk FOREIGN KEY (timetable_id) REFERENCES public.timetable(timetable_id);

CREATE POLICY rls_scheduling_slot ON "scheduling_slot" using (permission_check(resource_path, 'scheduling_slot')) with check (permission_check(resource_path, 'scheduling_slot'));
CREATE POLICY rls_scheduling_slot_restrictive ON "scheduling_slot" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'scheduling_slot')) with check (permission_check(resource_path, 'scheduling_slot'));

ALTER TABLE "scheduling_slot" ENABLE ROW LEVEL security;
ALTER TABLE "scheduling_slot" FORCE ROW LEVEL security;