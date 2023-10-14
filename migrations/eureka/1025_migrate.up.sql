ALTER TABLE ONLY public.assignments
    ADD COLUMN IF NOT EXISTS original_topic TEXT;

UPDATE public.assignments
SET original_topic = content -> 'topic_id'
WHERE original_topic IS NULL AND content IS NOT NULL AND content ->> 'topic_id' <> '';

----------------------------------------

CREATE TABLE IF NOT EXISTS public.topics_assignments (
     topic_id TEXT NOT NULL,
     assignment_id TEXT NOT NULL,
     display_order SMALLINT,
     created_at timestamp with time zone NOT NULL,
     updated_at timestamp with time zone NOT NULL,
     deleted_at timestamp with time zone
);

ALTER TABLE ONLY public.topics_assignments
    DROP CONSTRAINT IF EXISTS topics_assignments_pk,
    DROP CONSTRAINT IF EXISTS topics_assignments_assignment_fk;

ALTER TABLE ONLY public.topics_assignments
    ADD CONSTRAINT topics_assignments_pk PRIMARY KEY (topic_id, assignment_id),
    ADD CONSTRAINT topics_assignments_assignment_fk FOREIGN KEY (assignment_id) REFERENCES public.assignments(assignment_id);

-- migrate old data from table assignments to new table topics_assignments
INSERT INTO public.topics_assignments(topic_id, assignment_id, display_order, created_at, updated_at, deleted_at)
SELECT a.original_topic, a.assignment_id, a.display_order, a.created_at, a.updated_at, a.deleted_at
FROM public.assignments a
WHERE a.original_topic IS NOT NULL
ON CONFLICT ON CONSTRAINT topics_assignments_pk
    DO NOTHING;