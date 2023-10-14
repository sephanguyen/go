UPDATE
    quizzes
SET
    question = replace(question::text, '"FLASHCARD_LANGUAGE_CONFIG_ENG"', '"FLASHCARD_LANGUAGE_CONFIG_ENG", "LANGUAGE_CONFIG_ENG"')::jsonb;

UPDATE
    quizzes
SET
    question = replace(question::text, '"FLASHCARD_LANGUAGE_CONFIG_JP"', '"FLASHCARD_LANGUAGE_CONFIG_JP", "LANGUAGE_CONFIG_JP"')::jsonb;