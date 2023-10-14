CREATE OR REPLACE FUNCTION public.lo_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text)
    LANGUAGE sql STABLE
    AS $$
select  sa.student_id,
        sa.study_plan_id,
        sa.learning_material_id,
        sa.submission_id,
        sum(point)::smallint as graded_point,
        max(s.total_point)::smallint as total_point,
        'S'
from lo_submission_answer sa
join lo_submission s using (submission_id)
group by sa.student_id,
         sa.study_plan_id,
         sa.learning_material_id,
         sa.submission_id
$$;

CREATE OR REPLACE FUNCTION public.fc_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text)
    LANGUAGE sql STABLE
    AS $$
 select sa.student_id,
        sa.study_plan_id,
        sa.learning_material_id,
        sa.submission_id,
        sum(point)::smallint as graded_point,
        max(s.total_point)::smallint as total_point,
        'S'
from flash_card_submission_answer sa
join flash_card_submission s using (submission_id)
join get_student_completion_learning_material() clm on clm.student_id = sa.student_id and clm.study_plan_id = sa.study_plan_id and clm.learning_material_id = sa.learning_material_id
group by sa.student_id,
         sa.study_plan_id,
         sa.learning_material_id,
         sa.submission_id
$$;

CREATE OR REPLACE FUNCTION assignment_graded_score_v2()
    returns table
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                student_submission_id text,
                graded_point         smallint,
                total_point          smallint,
                status               text,
                passed               bool,
                created_at timestamptz
            )
    language sql
    stable
as
$$
select 
    ss.student_id,
    ss.study_plan_id,
    ss.learning_material_id,
    ss.student_submission_id,
    ssg.grade::smallint as graded_point,
    a.max_grade::smallint as total_point,
    ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at 
from student_submissions ss
join student_submission_grades ssg on ss.student_submission_id = ssg.student_submission_id
join assignment a using (learning_material_id)
where ssg.grade != -1 and ss.status = 'SUBMISSION_STATUS_RETURNED'
order by ss.student_id, ss.study_plan_id, ss.learning_material_id, ss.created_at;
$$;

CREATE OR REPLACE FUNCTION public.exam_lo_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text, result text, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
 select els.student_id,
    els.study_plan_id,
    els.learning_material_id,
    els.submission_id,
    --   when teacher manual grading, we should use with score from teacher
    (
        (el.review_option = 'EXAM_LO_REVIEW_OPTION_IMMEDIATELY' or (now()>check_study_plan_item_time(msp.updated_at, isp.updated_at, msp.end_date,isp.end_date)))::BOOLEAN::INT*sum(coalesce(elss.point, elsa.point))
    )::smallint as graded_point,
    els.total_point::smallint as total_point,
    els.status,
    els.result,
    els.created_at
from exam_lo_submission els
    join exam_lo_submission_answer elsa using (submission_id)
    left join exam_lo_submission_score elss using (submission_id, quiz_id)
    join exam_lo el on els.learning_material_id = el.learning_material_id 
    left join master_study_plan msp on msp.study_plan_id = els.study_plan_id and msp.learning_material_id = els.learning_material_id  
    left join individual_study_plan isp on isp.learning_material_id = els.learning_material_id and isp.study_plan_id = els.study_plan_id and isp.student_id = els.student_id 
    where els.status = 'SUBMISSION_STATUS_RETURNED'
group by els.submission_id,el.review_option,msp.updated_at,isp.updated_at,msp.end_date,isp.end_date
$$;

CREATE OR REPLACE FUNCTION public.task_assignment_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, student_submission_id text, graded_point smallint, total_point smallint, status text, passed boolean, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
select
    ss.student_id,
    ss.study_plan_id,
    ss.learning_material_id,
    ss.student_submission_id,
    ss.correct_score::smallint as graded_point,
    ss.total_score::smallint as total_point,
    ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at
from student_submissions ss
    join task_assignment ta using (learning_material_id)
where ss.correct_score > 0;
$$;

-- fix trigger lo_submission and lo_submission_answer

CREATE OR REPLACE FUNCTION public.migrate_to_lo_submission_and_answer_fnc()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $FUNCTION$
BEGIN
  -- insert to lo submission first
    IF EXISTS (
        SELECT 1 
        FROM public.learning_objective LO 
        WHERE LO.learning_material_id = NEW.learning_material_id and NEW.submission_history::text != '[]'::text
    )
    THEN
        INSERT INTO public.lo_submission (
            submission_id,
            student_id,
            study_plan_id,
            learning_material_id,
            shuffled_quiz_set_id,
            total_point,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        )
        VALUES (
            generate_ulid(),
            NEW.student_id,
            NEW.study_plan_id,
            NEW.learning_material_id,
            NEW.shuffled_quiz_set_id,
            COALESCE((SELECT SUM(point) FROM public.quizzes WHERE quizzes.deleted_at IS NULL AND quizzes.external_id = ANY(NEW.quiz_external_ids)), 0),
            NEW.created_at,
            NEW.updated_at,
            NEW.deleted_at,
            NEW.resource_path
        )
        ON CONFLICT ON CONSTRAINT shuffled_quiz_set_id_lo_submission_un DO UPDATE SET
            updated_at = EXCLUDED.updated_at,
            deleted_at = EXCLUDED.deleted_at,
            total_point = EXCLUDED.total_point;

  -- continue insert to lo answer 
        INSERT INTO public.lo_submission_answer(
        student_id,
        quiz_id,
        submission_id,
        study_plan_id,
        learning_material_id,
        shuffled_quiz_set_id,
        student_text_answer,
        correct_text_answer,
        student_index_answer,
        correct_index_answer,
        submitted_keys_answer,
        correct_keys_answer,
        point,
        is_correct,
        is_accepted,
        created_at,
        updated_at,
        deleted_at,
        resource_path
    )
    SELECT 
        sh.student_id,
        sh.quiz_id,
        ls.submission_id,
        ls.study_plan_id,
        ls.learning_material_id,
        sh.shuffled_quiz_set_id,
        sh.student_text_answer,
        sh.correct_text_answer,
        sh.student_index_answer,
        sh.correct_index_answer,
        sh.submitted_keys_answer,
        sh.correct_keys_answer,
        sh.point,
        sh.is_correct,
        sh.is_accepted,
        sh.created_at,
        sh.updated_at,
        sh.deleted_at,
        sh.resource_path
    FROM get_submission_history() AS sh
    JOIN lo_submission ls USING(shuffled_quiz_set_id)
    JOIN quizzes q ON q.external_id = sh.quiz_id
    WHERE ls.deleted_at IS NULL
        AND q.deleted_at IS NULL
        AND sh.shuffled_quiz_set_id = NEW.shuffled_quiz_set_id
    ON CONFLICT ON CONSTRAINT lo_submission_answer_pk DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        correct_text_answer = EXCLUDED.correct_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
        correct_index_answer = EXCLUDED.correct_index_answer,
        submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        correct_keys_answer = EXCLUDED.correct_keys_answer,
        point = EXCLUDED.point,
        is_correct = EXCLUDED.is_correct,
        is_accepted = EXCLUDED.is_accepted,
        updated_at = EXCLUDED.updated_at,
        deleted_at = EXCLUDED.deleted_at;
    END IF;
  
RETURN NULL;
END;
$FUNCTION$;

-- drop and create trigger

DROP TRIGGER IF EXISTS migrate_to_lo_submission_and_answer ON public.shuffled_quiz_sets;

CREATE TRIGGER migrate_to_lo_submission_and_answer
    AFTER UPDATE OF updated_at ON public.shuffled_quiz_sets
    FOR EACH ROW
    EXECUTE FUNCTION public.migrate_to_lo_submission_and_answer_fnc();

DROP TRIGGER IF EXISTS migrate_to_lo_submission ON public.shuffled_quiz_sets;
DROP TRIGGER IF EXISTS migrate_to_lo_submission_answer ON public.shuffled_quiz_sets;

-- create new func graded_score with new table and accepted score (task assignment + assignment + exam_lo)
CREATE OR REPLACE FUNCTION public.graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text)
    LANGUAGE sql STABLE
    AS $$
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_point,
    total_point,
    status
from lo_graded_score_v2()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_point,
    total_point,
    status
from fc_graded_score_v2()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    student_submission_id,
    graded_point,
    total_point,
    status
from assignment_graded_score_v2()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    student_submission_id,
    graded_point,
    total_point,
    status
from task_assignment_graded_score_v2()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_point,
    total_point,
    status
from exam_lo_graded_score_v2() 
$$;

CREATE OR REPLACE FUNCTION public.max_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, graded_point smallint, total_point smallint)
    LANGUAGE sql STABLE
    AS $$
select distinct on (student_id, 
    study_plan_id ,
    learning_material_id) student_id,
                          study_plan_id,
                          learning_material_id,
                          graded_point,
                          total_point
from
    graded_score_v2()
where total_point > 0
order by student_id, study_plan_id, learning_material_id, graded_point * 1.0 / total_point desc
$$;