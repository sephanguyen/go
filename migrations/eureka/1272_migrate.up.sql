ALTER TABLE quizzes ADD COLUMN IF NOT EXISTS label_type TEXT NULL DEFAULT 'QUIZ_LABEL_TYPE_NONE';

ALTER TABLE quizzes
ADD CONSTRAINT label_type_check CHECK(label_type IN (
'QUIZ_LABEL_TYPE_NONE',
'QUIZ_LABEL_TYPE_WITHOUT_LABEL',
'QUIZ_LABEL_TYPE_CUSTOM',
'QUIZ_LABEL_TYPE_NUMBER',
'QUIZ_LABEL_TYPE_TEXT_LOWERCASE',
'QUIZ_LABEL_TYPE_TEXT_UPPERCASE'
));
