CREATE TABLE IF NOT EXISTS public.time_slot (
    time_slot_id TEXT NOT NULL,
    time_slot_internal_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    CONSTRAINT time_slot_duration_location_id_unique UNIQUE(start_time, end_time, location_id),
    CONSTRAINT time_slot_internal_id_location_id_unique UNIQUE(time_slot_internal_id, location_id)
);

CREATE POLICY rls_time_slot ON "time_slot"
USING (permission_check(resource_path, 'time_slot')) WITH CHECK (permission_check(resource_path, 'time_slot'));
CREATE POLICY rls_time_slot_restrictive ON "time_slot" AS RESTRICTIVE
USING (permission_check(resource_path, 'time_slot'))WITH CHECK (permission_check(resource_path, 'time_slot'));

ALTER TABLE "time_slot" ENABLE ROW LEVEL security;
ALTER TABLE "time_slot" FORCE ROW LEVEL security;

ALTER TABLE public.working_hour ALTER column opening_time TYPE TIME USING opening_time::TIME WITHOUT TIME ZONE;

ALTER TABLE public.working_hour ALTER column closing_time TYPE TIME USING closing_time::TIME WITHOUT TIME ZONE;

ALTER TABLE public.academic_week ALTER column week_order SET NOT NULL;
