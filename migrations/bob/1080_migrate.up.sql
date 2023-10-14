CREATE TABLE IF NOT EXISTS public.flashcard_progressions (
    study_set_id TEXT NOT NULL,
    origin_study_set_id TEXT,
    student_id TEXT NOT NULL,
    study_plan_item_id TEXT NOT NULL,
    lo_id TEXT NOT NULL,
    quiz_external_ids TEXT[],
    studying_index INT,
    skipped_question_ids TEXT[],
    remembered_question_ids TEXT[],
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT flashcard_progressions_pk PRIMARY KEY (study_set_id),
    CONSTRAINT flashcard_progressions_student_id FOREIGN KEY (student_id) REFERENCES students(student_id),
    CONSTRAINT flashcard_progressions_un UNIQUE (study_set_id)
);
