-- create index for child tables, no need for parent table learning_material because it's alway empty
CREATE INDEX IF NOT EXISTS exam_lo_topic_id_idx ON public.exam_lo USING btree (topic_id);
CREATE INDEX IF NOT EXISTS assignment_topic_id_idx ON public."assignment" USING btree (topic_id);
CREATE INDEX IF NOT EXISTS offline_learning_topic_id_idx ON public.offline_learning USING btree (topic_id);