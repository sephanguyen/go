CREATE INDEX CONCURRENTLY IF NOT EXISTS shuffled_quiz_sets_learning_material_idx ON public.shuffled_quiz_sets USING btree (learning_material_id);
