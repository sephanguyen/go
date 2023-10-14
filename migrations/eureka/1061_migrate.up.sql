CREATE INDEX IF NOT EXISTS book_idx
  ON public.study_plans (book_id);

ALTER TABLE ONLY public.lo_study_plan_items
    ADD CONSTRAINT lo_study_plan_items_study_plan_item_id_un UNIQUE (study_plan_item_id);

ALTER TABLE ONLY public.assignment_study_plan_items
    ADD CONSTRAINT assignment_study_plan_items_study_plan_item_id_un UNIQUE (study_plan_item_id);
