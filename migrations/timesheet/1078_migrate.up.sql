CREATE TABLE IF NOT EXISTS public.timesheet_action_log(
  action_log_id text not null,
  timesheet_id text not null,  
  user_id text,
  is_system boolean not null default 'false',
  action text not null,
  executed_at timestamp with time zone NOT NULL,
  resource_path text default autofillresourcepath(),
  created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
  updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
  deleted_at timestamp with time zone,  
  CONSTRAINT pk__timesheet_action_log PRIMARY KEY(action_log_id));
  

CREATE POLICY rls_timesheet_action_log ON "timesheet_action_log"
    USING (permission_check (resource_path, 'timesheet_action_log'))
    WITH CHECK (permission_check (resource_path, 'timesheet_action_log'));

CREATE POLICY rls_timesheet_action_log_restrictive ON "timesheet_action_log" AS RESTRICTIVE FOR ALL TO PUBLIC 
USING (permission_check(resource_path, 'timesheet_action_log'))
WITH CHECK (permission_check(resource_path, 'timesheet_action_log'));

ALTER TABLE "timesheet_action_log" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "timesheet_action_log" FORCE ROW LEVEL SECURITY;