ALTER TABLE IF EXISTS public.timesheet
DROP CONSTRAINT IF EXISTS staff_id_location_id_timesheet_date__unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx__staff_id_location_id_timesheet_date ON public.timesheet
USING btree(staff_id,location_id,timesheet_date)
WHERE (deleted_at IS NULL);
