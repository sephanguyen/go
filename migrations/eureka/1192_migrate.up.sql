CREATE INDEX CONCURRENTLY IF NOT EXISTS exam_lo_submission_answer_submission_id_idx ON public.exam_lo_submission_answer USING btree (submission_id);
