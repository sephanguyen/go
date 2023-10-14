ALTER TABLE IF EXISTS public.timesheet_lesson_hours
ADD CONSTRAINT fk__timesheet_lesson_hours_timesheet_id FOREIGN KEY (timesheet_id) REFERENCES public.timesheet(timesheet_id);