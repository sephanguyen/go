DROP INDEX IF EXISTS resource_path_idx;
create index if not exists resource_path_idx
    on users (resource_path);


DROP INDEX IF EXISTS idx__timesheet_lesson_hours_timesheet_id;
CREATE INDEX idx__timesheet_lesson_hours_timesheet_id ON public.timesheet_lesson_hours USING btree(timesheet_id);
