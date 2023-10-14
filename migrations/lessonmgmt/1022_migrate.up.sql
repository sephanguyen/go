ALTER TABLE ONLY lesson_report_details
    ADD COLUMN IF NOT EXISTS "report_versions" INTEGER DEFAULT 0
