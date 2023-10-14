ALTER TABLE flashcard_progressions
    ADD COLUMN IF NOT EXISTS original_quiz_set_id TEXT;

DO $$
    BEGIN
        IF EXISTS(SELECT *
                  FROM information_schema.columns
                  WHERE table_name='flashcard_progressions' and column_name='origin_study_set_id')
        THEN
            ALTER TABLE flashcard_progressions RENAME COLUMN origin_study_set_id TO original_study_set_id;
        END IF;
    END $$;
