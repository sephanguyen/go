DROP INDEX IF EXISTS topic_assignments;
CREATE INDEX topic_assignments ON assignments (("content"->>'topic_id'));

DROP FUNCTION IF EXISTS find_assignment_by_topic_id;

CREATE OR REPLACE FUNCTION public.find_assignment_by_topic_id(ids text[]) RETURNS SETOF public.assignments
    LANGUAGE sql STABLE
    AS $$
 select * from assignments a where a."content"->>'topic_id' = any(ids::text[]);
$$;
