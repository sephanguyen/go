ALTER TABLE ONLY preset_study_plans_weekly
	ADD COLUMN IF NOT EXISTS start_date timestamp with time zone,
	ADD COLUMN IF NOT EXISTS end_date timestamp with time zone;


UPDATE courses SET course_type = 'COURSE_TYPE_CONTENT' WHERE course_type IS NULL;


ALTER TABLE ONLY courses ADD COLUMN IF NOT EXISTS preset_study_plan_id text;

ALTER TABLE courses_classes DROP COLUMN preset_study_plan_id;


DROP VIEW IF EXISTS public.teacher_by_school_id;
CREATE OR REPLACE VIEW public.teacher_by_school_id AS
 SELECT unnest(t.school_ids) AS school_id,
    t.teacher_id,
    t.created_at,
    t.updated_at,
	t.deleted_at
   FROM public.teachers t;

ALTER TABLE preset_study_plans DROP COLUMN version;

--should drop view before alter this column
DROP VIEW IF EXISTS public.lesson_schedules;
DROP VIEW IF EXISTS public.preset_study_plans_weekly_format;

ALTER TABLE ONLY preset_study_plans_weekly
	ALTER COLUMN lesson_id TYPE text;
