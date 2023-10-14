CREATE INDEX IF NOT EXISTS flash_card_submission_answer_study_plan_item_identity_idx ON public.flash_card_submission_answer(student_id, study_plan_id, learning_material_id);

CREATE OR REPLACE FUNCTION public.fc_answer_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, external_quiz_id text, is_accepted boolean, point integer)
    LANGUAGE sql STABLE
AS $$
select  student_id,
        study_plan_id,
        learning_material_id,
        submission_id,
        quiz_id,
        is_accepted,
        point
from flash_card_submission_answer
join get_student_completion_learning_material() using (student_id, study_plan_id, learning_material_id);
$$;
