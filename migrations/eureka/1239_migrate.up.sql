CREATE INDEX question_group_id_idx ON public.quizzes USING hash (question_group_id) WHERE ((question_group_id IS NOT NULL) AND (deleted_at IS NULL));
