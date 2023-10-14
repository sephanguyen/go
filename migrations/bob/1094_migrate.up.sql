CREATE TABLE IF NOT EXISTS public.flashcard_speeches (
    speech_id TEXT NOT NULL,
    sentence TEXT NOT NULL,
    settings TEXT[],
    link TEXT NOT NULL,
    type TEXT NOT NULL,
    quiz_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by TEXT,
    updated_by TEXT,
    CONSTRAINT flashcard_speeches_pk PRIMARY KEY (speech_id)
);
