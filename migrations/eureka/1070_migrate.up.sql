CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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

-- avoid duplicate quiz_sets
CREATE UNIQUE INDEX IF NOT EXISTS quiz_sets_approved_lo_id_idx ON public.quiz_sets (lo_id) WHERE (status = 'QUIZSET_STATUS_APPROVED' AND deleted_at IS NULL);

-- student_learning_time_by_daily
COMMENT ON COLUMN public.student_learning_time_by_daily.learning_time IS 'learning time in seconds unit';

CREATE SEQUENCE IF NOT EXISTS public.student_learning_time_by_daily_learning_time_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE IF EXISTS public.student_learning_time_by_daily_learning_time_id_seq OWNED BY public.student_learning_time_by_daily.learning_time_id;

ALTER TABLE IF EXISTS ONLY public.student_learning_time_by_daily ALTER COLUMN learning_time_id SET DEFAULT nextval('public.student_learning_time_by_daily_learning_time_id_seq'::regclass);

ALTER TABLE public.student_learning_time_by_daily DROP CONSTRAINT IF EXISTS student_learning_time_by_daily_un;
ALTER TABLE public.student_learning_time_by_daily DROP CONSTRAINT IF EXISTS student_learning_time_by_daily_pk;
ALTER TABLE ONLY public.student_learning_time_by_daily
    ADD CONSTRAINT student_learning_time_by_daily_un UNIQUE (student_id, day);
ALTER TABLE ONLY public.student_learning_time_by_daily
    ADD CONSTRAINT student_learning_time_by_daily_pk PRIMARY KEY (learning_time_id);