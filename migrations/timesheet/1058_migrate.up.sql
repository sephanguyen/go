DROP INDEX IF EXISTS idx__timesheet_date;
CREATE INDEX IF NOT EXISTS idx__timesheet_date ON public.timesheet (timesheet_date);

DROP INDEX IF EXISTS idx__other_working_hours_timesheet_id;
CREATE INDEX idx__other_working_hours_timesheet_id ON public.other_working_hours (timesheet_id);

DROP INDEX IF EXISTS idx__transportation_expense_timesheet_id;
CREATE INDEX idx__transportation_expense_timesheet_id ON public.transportation_expense (timesheet_id);