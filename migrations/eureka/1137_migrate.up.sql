DROP FUNCTION IF EXISTS find_a_quiz_in_quiz_set;

CREATE OR REPLACE FUNCTION public.find_a_quiz_in_quiz_set(quizId text, lOId text)
RETURNS SETOF quizzes

LANGUAGE sql STABLE AS $function$

SELECT quiz.* FROM public.quizzes AS quiz 
    WHERE 
        quiz.quiz_id = quizId
    AND quiz.deleted_at IS NULL
    AND (
        EXISTS (
            SELECT qs.* FROM public.quiz_sets AS qs
                where qs.lo_id = lOId
                AND qs.deleted_at IS NULL
                AND qs.quiz_external_ids && ARRAY[quiz.external_id]
        )
    );

$function$;
