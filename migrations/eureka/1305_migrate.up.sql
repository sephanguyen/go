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
    learning_material_id = current_learning_material_id,
    study_plan_id = current_study_plan_id
WHERE
    student_event_log_id = new.student_event_log_id;
END IF;
IF (new.event_type = 'learning_objective' and NEW.payload ->> 'event' = 'completed')
THEN
    IF exists (select 1 from flash_card where learning_material_id = current_learning_material_id) THEN
        update flash_card_submission set is_submitted = true, updated_at = now()    
        where student_id = new.student_id 
            and learning_material_id = current_learning_material_id
            and study_plan_id = current_study_plan_id 
            and shuffled_quiz_set_id = (select shuffled_quiz_set_id from shuffled_quiz_sets where session_id = new.payload ->> 'session_id' ORDER BY created_at desc limit 1);
    END IF;
    
    if exists (select 1 from learning_objective where learning_material_id = current_learning_material_id) THEN
        update lo_submission set is_submitted = true, updated_at = now()
        where student_id = new.student_id 
            and learning_material_id = current_learning_material_id
            and study_plan_id = current_study_plan_id 
            and shuffled_quiz_set_id = (select shuffled_quiz_set_id from shuffled_quiz_sets where session_id = new.payload ->> 'session_id' ORDER BY created_at desc limit 1);
    END IF;

	call upsert_highest_score(current_study_plan_id, current_learning_material_id, new.student_id, new.resource_path);
END IF;
RETURN NULL;
END;
$function$;