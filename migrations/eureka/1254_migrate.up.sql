create or replace function get_exam_lo_returned_scores() returns table (
        student_id text,
        study_plan_id text,
        learning_material_id text,
        submission_id text,
        graded_point smallint,
        total_point smallint,
        status text,
        result text,
        created_at timestamptz
    ) language sql stable as $$
select els.student_id,
    els.study_plan_id,
    els.learning_material_id,
    els.submission_id,
    --   when teacher manual grading, we should use with score from teacher
    (
        (els.status = 'SUBMISSION_STATUS_RETURNED')::BOOLEAN::INT * sum(coalesce(elss.point, elsa.point))
    )::smallint as graded_point,
    els.total_point::smallint as total_point,
    els.status,
    els.result,
    els.created_at
from exam_lo_submission els
    join exam_lo_submission_answer elsa using (submission_id)
    left join exam_lo_submission_score elss using (submission_id, quiz_id)
group by els.submission_id $$;

CREATE OR REPLACE FUNCTION graded_score() RETURNS TABLE(
        student_id text,
        study_plan_id text,
        learning_material_id text,
        submission_id text,
        graded_points smallint,
        total_points smallint,
        status text
    ) LANGUAGE sql STABLE AS $$
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_points,
    total_points,
    status
from lo_graded_score()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_points,
    total_points,
    status
from fc_graded_score()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    student_submission_id,
    graded_point,
    total_point,
    status
from get_assignment_scores()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    student_submission_id,
    graded_point,
    total_point,
    status
from get_task_assignment_scores()
union all
select student_id,
    study_plan_id,
    learning_material_id,
    submission_id,
    graded_point,
    total_point,
    status
from get_exam_lo_returned_scores() $$;