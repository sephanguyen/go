WITH InvalidTable AS (
     SELECT quiz_set_id
     FROM (
         SELECT DISTINCT lo_id, COUNT(*) AS total, MAX(updated_at) AS max_updated_at
         FROM public.quiz_sets
         WHERE status = 'QUIZSET_STATUS_APPROVED' AND deleted_at IS NULL
         GROUP BY lo_id
     ) AS tba
     JOIN public.quiz_sets AS tbb
     ON tba.lo_id = tbb.lo_id
 		AND tbb.updated_at < tba.max_updated_at
 		AND status = 'QUIZSET_STATUS_APPROVED'
 		AND deleted_at IS NULL
     WHERE total >= 2
 )

UPDATE public.quiz_sets as tb
SET status = 'QUIZSET_STATUS_DELETED'::TEXT, deleted_at = NOW()
WHERE EXISTS (
	SELECT 1
    FROM InvalidTable
	WHERE quiz_set_id = tb.quiz_set_id
);

CREATE UNIQUE INDEX IF NOT EXISTS quiz_sets_approved_lo_id_idx ON public.quiz_sets (lo_id) WHERE (status = 'QUIZSET_STATUS_APPROVED' AND deleted_at IS NULL);
