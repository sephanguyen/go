ALTER TABLE ONLY study_plan_items
	ADD COLUMN IF NOT EXISTS content_structure JSONB;
