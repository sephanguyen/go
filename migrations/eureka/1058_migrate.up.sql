DROP FUNCTION IF EXISTS find_quiz_by_lo_id;

CREATE OR REPLACE FUNCTION public.find_quiz_by_lo_id(id VARCHAR) RETURNS SETOF public.quizzes
    LANGUAGE sql STABLE
    AS $$
select q.* from quiz_sets qs, unnest(qs.quiz_external_ids) WITH ORDINALITY AS search_quiz_external_ids(quiz_external_id, ordinality)
                                  join quizzes q on (q.external_id::TEXT = search_quiz_external_ids.quiz_external_id )
where qs.deleted_at IS NULL AND q.deleted_at IS NULL and qs.lo_id = id
order by search_quiz_external_ids.ordinality ASC
    $$;
