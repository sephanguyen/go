ALTER TABLE IF EXISTS lesson_members
    ADD COLUMN IF NOT EXISTS attendance_notice TEXT,
    ADD COLUMN IF NOT EXISTS attendance_reason TEXT,
    ADD COLUMN IF NOT EXISTS attendance_note TEXT;