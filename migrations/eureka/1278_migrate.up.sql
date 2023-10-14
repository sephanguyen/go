DROP FUNCTION IF EXISTS public.fc_answer_v2();

CREATE OR REPLACE FUNCTION public.fc_answer_v2() 
    returns table
        (
            student_id              text, 
            study_plan_id           text, 
            learning_material_id    text, 
            submission_id           text, 
            external_quiz_id        text, 
            is_accepted             boolean, 
            point                   integer, 
            total_point             integer
        )
    LANGUAGE sql STABLE
AS $$
select  sa.student_id,
        sa.study_plan_id,
        sa.learning_material_id,
        sa.submission_id,
        sa.quiz_id,
        sa.is_accepted,
        point,
        s.total_point
from flash_card_submission_answer sa
join flash_card_submission s using(submission_id)
join get_student_completion_learning_material() clm on
    clm.student_id = sa.student_id and clm.study_plan_id = sa.study_plan_id and clm.learning_material_id = sa.learning_material_id
$$;

CREATE OR REPLACE FUNCTION public.lo_answer_v2()
    returns table
        (
            student_id           text,
            study_plan_id        text,
            learning_material_id text,
            submission_id        text,
            external_quiz_id     text,
            is_accepted          bool,
            point                integer,
            total_point          integer
        )
    LANGUAGE sql STABLE
AS
$$
select  sa.student_id,
        sa.study_plan_id,
        sa.learning_material_id,
        sa.submission_id,
        sa.quiz_id,
        sa.is_accepted,
        point,
        s.total_point
from lo_submission_answer sa
join lo_submission s using (submission_id)
$$;
