
DROP FUNCTION IF EXISTS find_quiz_by_school_id;
DROP FUNCTION IF EXISTS find_quiz_by_lo_id;

CREATE OR REPLACE FUNCTION public.find_quiz_by_lo_id(lo_id VARCHAR) RETURNS SETOF public.quizzes
    LANGUAGE sql STABLE
    AS $$
    select q.* from quiz_sets qs join quizzes q on (q.external_id::TEXT = ANY(qs.quiz_external_ids::text[])) where qs.lo_id = lo_id ORDER by q.external_id ASC, q.created_at DESC;
$$;

