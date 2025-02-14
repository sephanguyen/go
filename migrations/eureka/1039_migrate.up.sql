ALTER TABLE IF EXISTS study_plans
    ADD COLUMN IF NOT EXISTS book_id TEXT,
    ADD COLUMN IF NOT EXISTS "status" TEXT DEFAULT 'STUDY_PLAN_STATUS_ACTIVE'::TEXT,
    ADD COLUMN IF NOT EXISTS track_school_progress BOOLEAN DEFAULT false,
    DROP CONSTRAINT IF EXISTS study_plan_status_check,
    ADD CONSTRAINT study_plan_status_check CHECK (("status" = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::TEXT, 'STUDY_PLAN_STATUS_ACTIVE'::TEXT, 'STUDY_PLAN_STATUS_ARCHIVED'::TEXT]))),
    ADD COLUMN IF NOT EXISTS grades INTEGER[] DEFAULT '{}'::INTEGER[];