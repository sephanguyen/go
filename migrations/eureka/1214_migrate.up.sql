DROP FUNCTION IF EXISTS public.list_individual_study_plan_item();
CREATE OR REPLACE FUNCTION public.list_individual_study_plan_item() RETURNS TABLE (
        study_plan_id text,
        learning_material_id text,
        student_id text,
        available_from timestamp with time zone,
        available_to timestamp with time zone,
        start_date timestamp with time zone,
        end_date timestamp with time zone,
        status text,
        school_date timestamp with time zone,
        completed_at timestamp with time zone,
        lm_display_order smallint,
        scorce smallint,
        type text
    ) LANGUAGE sql STABLE AS $$
select distinct on (study_plan_id, learning_material_id, student_id)
    isp.study_plan_id,
    isp.learning_material_id,
    isp.student_id,
    isp.available_from,
    isp.available_to,
    isp.start_date,
    isp.end_date,
    isp.status,
    isp.school_date,
    gsl.completed_at,
    isp.lm_display_order,
	coalesce((gs.graded_points * 1.0 / gs.total_points) * 100, null)::smallint AS scorce,
    lm.type
FROM list_available_learning_material() AS isp
INNER JOIN course_study_plans csp using (study_plan_id)
INNER JOIN learning_material lm using (learning_material_id)
LEFT JOIN get_student_completion_learning_material() gsl using(student_id, study_plan_id, learning_material_id)
LEFT JOIN max_graded_score() gs using (student_id, study_plan_id, learning_material_id)
where
	csp.deleted_at IS NULL
	AND
    isp.status = 'STUDY_PLAN_ITEM_STATUS_ACTIVE' $$;
