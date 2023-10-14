ALTER TABLE ONLY public.student_entryexit_records 
  DROP CONSTRAINT IF EXISTS check_touch_events;

ALTER TABLE ONLY public.student_entryexit_records 
  DROP COLUMN IF EXISTS touch_event,
  DROP COLUMN IF EXISTS touched_at;

ALTER TABLE ONLY public.student_entryexit_records 
  ADD COLUMN IF NOT EXISTS entry_at timestamp with time zone NOT NULL,
  ADD COLUMN IF NOT EXISTS exit_at timestamp with time zone;