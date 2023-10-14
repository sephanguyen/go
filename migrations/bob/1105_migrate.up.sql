ALTER TABLE IF EXISTS flashcard_speeches DROP COLUMN IF EXISTS settings;
ALTER TABLE IF EXISTS flashcard_speeches ADD COLUMN settings JSONB;
