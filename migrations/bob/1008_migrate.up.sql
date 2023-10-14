ALTER TABLE ONLY courses
	ADD COLUMN IF NOT EXISTS course_type text,
	ADD COLUMN IF NOT EXISTS start_date timestamp with time zone,
	ADD COLUMN IF NOT EXISTS end_date timestamp with time zone,
	ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone,
	ADD COLUMN IF NOT EXISTS teacher_ids text[];

ALTER TABLE ONLY preset_study_plans 
	ADD COLUMN IF NOT EXISTS version smallint;

ALTER TABLE ONLY courses_classes
	ADD COLUMN IF NOT EXISTS preset_study_plan_id smallint;

ALTER TABLE ONLY preset_study_plans_weekly
	ADD COLUMN IF NOT EXISTS lesson_id smallint;

CREATE TABLE IF NOT EXISTS public.lessons (
    lesson_id text,
	teacher_id text,
	course_id text,
	created_at timestamp with time zone NOT NULL,
	updated_at timestamp with time zone NOT NULL,
	deleted_at timestamp with time zone
);

ALTER TABLE public.courses DROP CONSTRAINT IF EXISTS course_type_check;

ALTER TABLE ONLY courses
ADD CONSTRAINT course_type_check CHECK ((course_type = ANY (ARRAY['COURSE_TYPE_CONTENT'::text, 'COURSE_TYPE_LIVE'::text])));