CREATE TABLE IF NOT EXISTS shuffled_quiz_sets (
    shuffled_quiz_set_id TEXT,
    original_quiz_set_id TEXT,
    quiz_external_ids TEXT[],
    status TEXT,
    random_seed TEXT,
    updated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ NULL,

    PRIMARY KEY (shuffled_quiz_set_id),
    CONSTRAINT fk_original_quiz_set FOREIGN KEY (original_quiz_set_id) REFERENCES public.quiz_sets(quiz_set_id)
);

CREATE INDEX IF NOT EXISTS shuffled_quiz_original_quiz_set_id_idx ON public.shuffled_quiz_sets (original_quiz_set_id);
