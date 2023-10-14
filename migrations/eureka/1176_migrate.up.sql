create or replace function get_task_assignment_scores()
    returns table
            (
                student_id              text,
                study_plan_id           text,
                learning_material_id    text,
                student_submission_id   text,
                graded_point            smallint,
                total_point             smallint,
                status                  text,
                passed                  bool,
                created_at              timestamptz
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
    ss.correct_score::smallint as graded_point,
        ss.total_score::smallint as total_point,
        ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at
from student_submissions ss
    join task_assignment ta using (learning_material_id);
$$;
