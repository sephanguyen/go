CREATE TABLE IF NOT EXISTS public.topics_learning_objectives (
     topic_id TEXT NOT NULL,
     lo_id TEXT NOT NULL,
     display_order smallint,
     updated_at timestamp with time zone NOT NULL,
     created_at timestamp with time zone NOT NULL,
     deleted_at timestamp with time zone
);

ALTER TABLE ONLY public.topics_learning_objectives
    DROP CONSTRAINT IF EXISTS topics_learning_objectives_pk,
    DROP CONSTRAINT IF EXISTS topics_learning_objectives_lo_fk,
    DROP CONSTRAINT IF EXISTS topics_learning_objectives_topic_fk;

ALTER TABLE ONLY public.topics_learning_objectives
    ADD CONSTRAINT topics_learning_objectives_pk PRIMARY KEY (topic_id, lo_id),
    ADD CONSTRAINT topics_learning_objectives_lo_fk FOREIGN KEY (lo_id) REFERENCES public.learning_objectives(lo_id),
    ADD CONSTRAINT topics_learning_objectives_topic_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id);

-- migrate old data from table topics and learning_objectives to new table topics_learning_objectives
INSERT INTO public.topics_learning_objectives(topic_id, lo_id, display_order, updated_at, created_at, deleted_at)
SELECT lo.topic_id, lo.lo_id, lo.display_order, lo.updated_at, lo.created_at, lo.deleted_at
FROM public.learning_objectives lo
JOIN topics
ON lo.topic_id = topics.topic_id AND topics.deleted_at IS NULL
ON CONFLICT ON CONSTRAINT topics_learning_objectives_pk
DO NOTHING;
