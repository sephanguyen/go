create or replace function get_assignment_scores()
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
    ss.correct_score::smallint as graded_point,
    ss.total_score::smallint as total_point,
    ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at 
from student_submissions ss
    join assignment a using (learning_material_id)
order by ss.student_id, ss.study_plan_id, ss.learning_material_id, ss.created_at;
$$;

create or replace function get_max_assignment_scores()
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
                total_attempts       smallint,
                created_at           timestamptz
            )
    language sql
    stable
as
$$
select distinct on (gas.student_id,gas.study_plan_id,gas.learning_material_id)
    gas.student_id,
    gas.study_plan_id,
    gas.learning_material_id,
    gas.student_submission_id,
    gas.graded_point,
    gas.total_point,
    gas.status,
    gas.passed,
     count(*)
    over (partition by gas.student_id, gas.study_plan_id, gas.learning_material_id)::smallint as total_attempts,
    gas.created_at 
from get_assignment_scores() gas
order by gas.student_id, gas.study_plan_id, gas.learning_material_id,gas.graded_point * 1.0 / coalesce(nullif(gas.total_point,0),1) desc;
$$;