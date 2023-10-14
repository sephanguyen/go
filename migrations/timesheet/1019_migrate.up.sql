ALTER TABLE IF EXISTS timesheet
ADD CONSTRAINT staff_id_location_id_timesheet_date__unique UNIQUE (staff_id,location_id,timesheet_date);