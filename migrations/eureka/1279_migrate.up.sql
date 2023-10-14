create or replace function get_assignment_scores_v2()
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
where ssg.grade != -1
order by ss.student_id, ss.study_plan_id, ss.learning_material_id, ss.created_at;
$$;