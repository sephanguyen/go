ALTER TABLE ONLY timesheet_config ALTER COLUMN created_at SET DEFAULT (now() at time zone 'utc');
ALTER TABLE ONLY timesheet_config ALTER COLUMN updated_at SET DEFAULT (now() at time zone 'utc');