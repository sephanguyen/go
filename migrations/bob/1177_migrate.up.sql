ALTER TABLE public.student_learning_time_by_daily 
ADD IF NOT EXISTS assignment_learning_time int4 NOT NULL DEFAULT 0;

ALTER TABLE public.student_learning_time_by_daily 
ADD IF NOT EXISTS assignment_submission_ids text[] NULL;

ALTER TABLE public.student_learning_time_by_daily 
ALTER COLUMN sessions DROP NOT NULL;