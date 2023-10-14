--
-- Name: topics_chapter_id_idx; Type: INDEX; Schema: public
--

CREATE INDEX CONCURRENTLY IF NOT EXISTS topics_chapter_id_idx ON public.topics USING btree (chapter_id);
