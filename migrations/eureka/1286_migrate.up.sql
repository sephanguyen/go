CREATE OR REPLACE procedure upsert_highest_score(current_study_plan_id text, current_learning_material_id text, current_student_id text, current_resource_path text)
 LANGUAGE plpgsql
AS $$
declare tmp record;
BEGIN 
	select  
		coalesce(graded_point,0) as max_score,
        coalesce(total_point,0) as total_score, 
        coalesce((graded_point * 1.0 / total_point) * 100, 0)::smallint max_percentage 
	from max_graded_score_v2()
	where study_plan_id = current_study_plan_id 
	and learning_material_id = current_learning_material_id
	and student_id = current_student_id
	into tmp;
	insert into max_score_submission (study_plan_id, learning_material_id, student_id, max_score, total_score, max_percentage, created_at, updated_at, deleted_at, resource_path)
	values(current_study_plan_id, current_learning_material_id, current_student_id,tmp.max_score, tmp.total_score, tmp.max_percentage, now(), now(), null, current_resource_path) 
	ON CONFLICT ON constraint max_score_submission_study_plan_item_identity_pk 
	do update set max_score = tmp.max_score,
				  total_score = tmp.total_score,
                  max_percentage = tmp.max_percentage,
				  updated_at = now();
end; 
$$;

CREATE OR REPLACE FUNCTION public.trigger_student_event_logs_fill_new_identity_fn()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $function$
DECLARE current_study_plan_id text;
        current_learning_material_id text;
BEGIN 
IF new.event_type = ANY(ARRAY[
    'study_guide_finished',
    'video_finished',
    'learning_objective',
    'quiz_answer_selected'
    ]) THEN
    current_study_plan_id =(
        SELECT
        COALESCE(sp.master_study_plan_id, sp.study_plan_id)
        FROM
        study_plan_items spi
        JOIN study_plans sp ON spi.study_plan_id = sp.study_plan_id
        WHERE
        spi.study_plan_item_id = new.payload ->> 'study_plan_item_id'
    );
    current_learning_material_id = new.payload->>'lo_id';
UPDATE
    student_event_logs sel
SET
    study_plan_item_id = (new.payload ->> 'study_plan_item_id'),
    learning_material_id = current_learning_material_id
WHERE
    student_event_log_id = new.student_event_log_id;
END IF;
IF (new.event_type = 'learning_objective' and NEW.payload ->> 'event' = 'completed')
THEN
	call upsert_highest_score(current_study_plan_id, current_learning_material_id, new.student_id, new.resource_path);
END IF;
RETURN NULL;
END;
$function$;
