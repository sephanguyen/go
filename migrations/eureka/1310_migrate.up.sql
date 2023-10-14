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
where ss.correct_score > 0 and ss.deleted_at is null;
$$;