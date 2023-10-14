ALTER TABLE ONLY study_plan_items
	ADD COLUMN IF NOT EXISTS copy_study_plan_item_id text;
ALTER TABLE public.course_classes ADD course_class_id text NOT NULL;
