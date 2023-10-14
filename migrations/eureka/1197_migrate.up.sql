CREATE INDEX IF NOT EXISTS master_study_plan_study_plan_id_idx ON public.master_study_plan USING btree (study_plan_id);

CREATE INDEX IF NOT EXISTS master_study_plan_lm_id_idx ON public.master_study_plan USING btree (learning_material_id);

CREATE INDEX IF NOT EXISTS course_study_plans_course_id_idx ON public.course_study_plans USING btree (course_id);

CREATE INDEX IF NOT EXISTS course_study_plans_study_plan_id_idx ON public.course_study_plans USING btree (study_plan_id);

CREATE INDEX IF NOT EXISTS students_current_grade_idx ON public.students USING btree (current_grade);
