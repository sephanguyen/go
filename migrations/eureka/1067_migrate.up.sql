ALTER TABLE public.student_submissions ADD IF NOT EXISTS complete_date timestamptz NULL;
ALTER TABLE public.student_submissions ADD IF NOT EXISTS duration int4 NULL;
ALTER TABLE public.student_submissions ADD IF NOT EXISTS correct_score float4 NULL;
ALTER TABLE public.student_submissions ADD IF NOT EXISTS total_score float4 NULL;
ALTER TABLE public.student_submissions ADD IF NOT EXISTS understanding_level text NULL;

ALTER TABLE public.student_latest_submissions ADD IF NOT EXISTS complete_date timestamptz NULL;
ALTER TABLE public.student_latest_submissions ADD IF NOT EXISTS duration int4 NULL;
ALTER TABLE public.student_latest_submissions ADD IF NOT EXISTS correct_score float4 NULL;
ALTER TABLE public.student_latest_submissions ADD IF NOT EXISTS total_score float4 NULL;
ALTER TABLE public.student_latest_submissions ADD IF NOT EXISTS understanding_level text NULL;
