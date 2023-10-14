CREATE INDEX IF NOT EXISTS idx__other_working_hours_timesheet_id ON public.other_working_hours USING hash (timesheet_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx__transportation_expense_timesheet_id ON public.transportation_expense USING hash (timesheet_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx__location_name ON public.locations USING GIN (name gin_trgm_ops);