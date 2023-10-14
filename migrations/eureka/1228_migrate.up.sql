ALTER TABLE IF EXISTS public.assignments
    ADD COLUMN IF NOT EXISTS topic_id TEXT;

CREATE INDEX IF NOT EXISTS assignments_topic_id_idx ON public.assignments (topic_id);

-- Migrate old data
UPDATE
    public.assignments
SET
    topic_id = content ->> 'topic_id'
WHERE
    content ->> 'topic_id' IS NOT NULL
